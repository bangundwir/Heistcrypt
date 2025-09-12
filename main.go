package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
    "strings"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/widget"

	"io"
	"encoding/binary"
	"sync/atomic"
	"github.com/bangundwir/HadesCrypt/internal/archiver"
	"github.com/bangundwir/HadesCrypt/internal/config"
	"github.com/bangundwir/HadesCrypt/internal/cryptoengine"
	"github.com/bangundwir/HadesCrypt/internal/keyfiles"
	pw "github.com/bangundwir/HadesCrypt/internal/password"
	uiutil "github.com/bangundwir/HadesCrypt/internal/ui"
)

// version is set at build time via -ldflags "-X main.version=<ver>"
var version string

type AppState struct {
	selectedPath        string
	selectedPaths       []string
	password            string
	confirmPassword     string
	comments            string
	strengthBar         *widget.ProgressBar
	strengthLabel       *widget.Label
	progressBar         *widget.ProgressBar
	statusLabel         *widget.Label
	fileInfoLabel       *widget.Label
	dragDropLabel       *widget.Label
	commentsEntry       *widget.Entry
	passwordEntry       *widget.Entry
	confirmPasswordEntry *widget.Entry
	passwordMatchLabel  *widget.Label
	config              *config.Config
	
	// Keyfiles
	keyfileManager   *keyfiles.KeyfileManager
	keyfilesList     *widget.List
	keyfilesLabel    *widget.Label
	
	// Encryption mode
	encryptionMode   cryptoengine.EncryptionMode
	
	// Advanced options
	deleteAfter      bool
	useKeyfiles      bool
	cancelRequested  atomic.Bool
	paranoidMode     bool
	reedSolomon      bool
	forceDecrypt     bool
	splitOutput      bool
	splitSize        int
	splitUnit        string
	compressFiles    bool
	deniabilityMode  bool
	recursiveMode    bool
}

// computeMixedSelectionSize walks selectedPaths computing total bytes that will be processed.
// For folders: sums all regular files (excluding already encrypted outputs) according to mode.
func (s *AppState) computeMixedSelectionSize(recursive bool) (int64, map[string]int64) {
	sizes := make(map[string]int64)
	var total int64
	for _, p := range s.selectedPaths {
		fi, err := os.Stat(p); if err != nil { continue }
		if fi.IsDir() {
			// sum all eligible files inside
			filepath.Walk(p, func(sp string, info os.FileInfo, err error) error {
				if err != nil || info == nil { return nil }
				if info.IsDir() { return nil }
				low := strings.ToLower(sp)
				if strings.HasSuffix(low, ".hadescrypt") || strings.HasSuffix(low, ".heistcrypt") || strings.HasSuffix(low, ".gpg") || strings.HasSuffix(low, ".pgp") { return nil }
				total += info.Size()
				sizes[p] += info.Size()
				return nil
			})
		} else if fi.Mode().IsRegular() {
			total += fi.Size()
			sizes[p] = fi.Size()
		}
	}
	return total, sizes
}

func main() {
	// version is injected via -X main.version at build time (see dist/windows/build.bat)
	// default to VERSION file or "dev"
	if version == "" {
		if data, err := os.ReadFile("VERSION"); err == nil {
			version = strings.TrimSpace(string(data))
		} else {
			version = "dev"
		}
	}
    application := app.NewWithID("hadescrypt")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use default config if loading fails
		cfg = config.DefaultConfig()
	}

	// Set theme based on config
	if cfg.Theme == "light" {
		application.Settings().SetTheme(theme.LightTheme())
	} else {
    application.Settings().SetTheme(theme.DarkTheme())
	}

	w := application.NewWindow(fmt.Sprintf("HadesCrypt v%s 🔱 — Lock your secrets, rule your data.", version))
	w.Resize(fyne.NewSize(cfg.WindowWidth, cfg.WindowHeight))
	w.CenterOnScreen()

	state := &AppState{
		config:         cfg,
		keyfileManager: keyfiles.NewKeyfileManager(),
		encryptionMode: cryptoengine.ModeAES256GCM,
		deleteAfter:    true, // Default to delete source files
	}
	state.setupUI(w)

	// Save window size on close
	w.SetCloseIntercept(func() {
		cfg.WindowWidth = w.Content().Size().Width
		cfg.WindowHeight = w.Content().Size().Height
		cfg.Save() // Save config on exit
		w.Close()
	})

	w.ShowAndRun()
}

func (s *AppState) setupUI(w fyne.Window) {
	// Header
	header := widget.NewLabelWithStyle("HadesCrypt 🔱", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	tagline := widget.NewLabelWithStyle("Lock your secrets, rule your data.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Drag & Drop Area (supports files and folders)
	s.dragDropLabel = widget.NewLabelWithStyle("[ Drag & Drop files or folders here ]", fyne.TextAlignCenter, fyne.TextStyle{})
	s.fileInfoLabel = widget.NewLabel("")
	
	dragDropContainer := container.NewVBox(
		s.dragDropLabel,
		s.fileInfoLabel,
	)
	
	// Create a card-like container for drag & drop
	dragDropCard := container.NewBorder(
		widget.NewSeparator(),
		widget.NewSeparator(),
		widget.NewSeparator(),
		widget.NewSeparator(),
		container.NewPadded(dragDropContainer),
	)

	// File / Folder selection buttons
	selectFileBtn := widget.NewButton("Select File", func() {
		s.showFileDialog(w)
	})
	selectFolderBtn := widget.NewButton("Select Folder", func() {
		s.showFolderDialog(w)
	})
	selectButtons := container.NewHBox(selectFileBtn, selectFolderBtn)

    // Password controls
	s.passwordEntry = widget.NewPasswordEntry()
	s.passwordEntry.SetPlaceHolder("Enter password…")
	s.passwordEntry.OnChanged = func(text string) {
		s.password = text
		s.updateStrength(text)
		s.validatePasswordMatch()
	}

	s.confirmPasswordEntry = widget.NewPasswordEntry()
	s.confirmPasswordEntry.SetPlaceHolder("Confirm password…")
	s.confirmPasswordEntry.OnChanged = func(text string) {
		s.confirmPassword = text
		s.validatePasswordMatch()
	}

	// Password match indicator
	s.passwordMatchLabel = widget.NewLabel("")

    genBtn := widget.NewButton("Generate", func() {
		s.showPasswordGeneratorDialog(w)
	})

	// Encryption mode selection
	encryptionModeSelect := widget.NewSelect(
		[]string{
			"AES-256-GCM", 
			"ChaCha20-Poly1305", 
			"Paranoid (AES-256 + ChaCha20)",
			"🛡️ Post-Quantum: Kyber-768",
			"🛡️ Post-Quantum: Dilithium-3",
			"🛡️ Post-Quantum: SPHINCS+",
			"🔐 GnuPG/OpenPGP (Standard)",
		},
		func(selected string) {
			switch selected {
			case "AES-256-GCM":
				s.encryptionMode = cryptoengine.ModeAES256GCM
			case "ChaCha20-Poly1305":
				s.encryptionMode = cryptoengine.ModeChaCha20
			case "Paranoid (AES-256 + ChaCha20)":
				s.encryptionMode = cryptoengine.ModeParanoid
			case "🛡️ Post-Quantum: Kyber-768":
				s.encryptionMode = cryptoengine.ModePostQuantumKyber768
			case "🛡️ Post-Quantum: Dilithium-3":
				s.encryptionMode = cryptoengine.ModePostQuantumDilithium3
			case "🛡️ Post-Quantum: SPHINCS+":
				s.encryptionMode = cryptoengine.ModePostQuantumSPHINCS
			case "🔐 GnuPG/OpenPGP (Standard)":
				s.encryptionMode = cryptoengine.ModeGnuPG
			}
		},
	)
	encryptionModeSelect.SetSelected("AES-256-GCM")

	// Keyfiles section
	s.keyfilesLabel = widget.NewLabel("No keyfiles selected")
	s.keyfilesList = widget.NewList(
		func() int {
			return s.keyfileManager.Count()
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("keyfile")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			paths := s.keyfileManager.GetPaths()
			if id < len(paths) {
				obj.(*widget.Label).SetText(filepath.Base(paths[id]))
			}
		},
	)
	s.keyfilesList.Resize(fyne.NewSize(400, 100))

	addKeyfileBtn := widget.NewButton("Add Keyfile", func() {
		s.showKeyfileDialog(w)
	})

	generateKeyfileBtn := widget.NewButton("Generate Keyfile", func() {
		s.showGenerateKeyfileDialog(w)
	})

	clearKeyfilesBtn := widget.NewButton("Clear All", func() {
		s.keyfileManager.Clear()
		s.updateKeyfilesDisplay()
	})

	keyfileButtons := container.NewHBox(addKeyfileBtn, generateKeyfileBtn, clearKeyfilesBtn)

	// Comments field
	s.commentsEntry = widget.NewMultiLineEntry()
	s.commentsEntry.SetPlaceHolder("Optional comments (not encrypted)...")
	s.commentsEntry.Resize(fyne.NewSize(400, 60))
	s.commentsEntry.OnChanged = func(text string) {
		s.comments = text
	}

	// Password strength meter
	s.strengthBar = widget.NewProgressBar()
	s.strengthBar.Min = 0
	s.strengthBar.Max = 1
	s.strengthLabel = widget.NewLabel("Strength: —")

    // Action buttons
	encryptBtn := widget.NewButton("🔒 Encrypt", func() {
		s.doEncrypt(w)
	})
	encryptBtn.Importance = widget.HighImportance

	decryptBtn := widget.NewButton("🔓 Decrypt", func() {
		s.doDecrypt(w)
	})

	// Progress and status
	s.progressBar = widget.NewProgressBar()
	s.progressBar.Min = 0
	s.progressBar.Max = 1
	s.statusLabel = widget.NewLabel("Status: Ready")

	// Advanced options
	advanced := s.buildAdvancedPanel()

	// Layout
	passwordRow := container.NewBorder(
        nil, nil,
        widget.NewLabel("Password:"),
		genBtn,
        s.passwordEntry,
    )

	confirmPasswordRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Confirm:"),
		s.passwordMatchLabel,
		s.confirmPasswordEntry,
	)

	strengthRow := container.NewBorder(
		nil, nil,
        widget.NewLabel("Strength:"),
		s.strengthLabel,
		s.strengthBar,
	)

	encryptionRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Encryption:"),
		nil,
		encryptionModeSelect,
	)

	keyfilesSection := container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Keyfiles:"), s.keyfilesLabel, nil),
		s.keyfilesList,
		keyfileButtons,
	)

	commentsRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Comments:"),
		nil,
		s.commentsEntry,
	)

	actionsRow := container.NewHBox(
		encryptBtn,
		decryptBtn,
		widget.NewButton("Cancel", func(){
			if !s.cancelRequested.Load() {
				s.cancelRequested.Store(true)
				s.statusLabel.SetText("Cancel requested…")
			}
		}),
	)

	progressRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Progress:"),
		nil,
		s.progressBar,
	)

	content := container.NewVBox(
		container.NewPadded(container.NewVBox(header, tagline)),
		widget.NewSeparator(),
		container.NewPadded(dragDropCard),
		container.NewPadded(selectButtons),
		widget.NewSeparator(),
		container.NewPadded(passwordRow),
		container.NewPadded(confirmPasswordRow),
		container.NewPadded(strengthRow),
		container.NewPadded(encryptionRow),
		container.NewPadded(keyfilesSection),
		container.NewPadded(commentsRow),
		widget.NewSeparator(),
		container.NewPadded(actionsRow),
		container.NewPadded(progressRow),
		container.NewPadded(s.statusLabel),
		widget.NewSeparator(),
		advanced,
	)

	// Set up drag and drop
	w.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 { return }
		if len(uris) == 1 {
			s.setSelectedFile(uris[0].Path())
			return
		}
		var paths []string
		for _, u := range uris { paths = append(paths, u.Path()) }
		if len(paths) == 1 { s.setSelectedFile(paths[0]); return }
		s.setSelectedFiles(paths)
	})

	w.SetContent(container.NewScroll(content))
}

func (s *AppState) showFileDialog(w fyne.Window) {
        fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
            if err != nil {
                dialog.ShowError(err, w)
                return
            }
            if rc == nil {
                return
            }
		defer rc.Close()
		
		path := rc.URI().Path()
		s.setSelectedFile(path)
	}, w)
	
	// Allow both files and folders
        fd.SetFilter(nil)
        fd.Show()
}

// showFolderDialog opens a folder selection dialog for selecting directories
func (s *AppState) showFolderDialog(w fyne.Window) {
	fd := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if list == nil {
			return
		}
		path := list.Path()
		s.setSelectedFile(path)
	}, w)
	fd.Show()
}

func (s *AppState) setSelectedFile(path string) {
	// Single selection resets multi selection
	s.selectedPath = path
	s.selectedPaths = nil
	s.updateFileInfo()
}

// setSelectedFiles sets multiple file selections (files only, no directories yet)
func (s *AppState) setSelectedFiles(paths []string) {
    s.selectedPath = ""
    s.selectedPaths = paths
    s.updateFileInfo()
}

func (s *AppState) updateFileInfo() {
	if s.selectedPath == "" && len(s.selectedPaths) == 0 {
		s.dragDropLabel.SetText("[ Drag & Drop your files here ]")
		s.fileInfoLabel.SetText("")
		s.commentsEntry.SetText("")
		return
	}

	if len(s.selectedPaths) > 0 {
		// Mixed multi selection summary
		var totalSize int64
		files := 0
		folders := 0
		for _, p := range s.selectedPaths {
			if fi, err := os.Stat(p); err == nil {
				if fi.IsDir() { folders++ } else if fi.Mode().IsRegular() { files++; totalSize += fi.Size() }
			}
		}
		icon := "📄"
		if folders > 0 { icon = "📂" }
		label := fmt.Sprintf("%s %d item(s) (%d file(s), %d folder(s))", icon, len(s.selectedPaths), files, folders)
		s.dragDropLabel.SetText(label)
		s.fileInfoLabel.SetText(fmt.Sprintf("Regular file bytes (pre-archive): %s", uiutil.HumanBytes(totalSize)))
		return
	}

	fileName := filepath.Base(s.selectedPath)
	
	// Check if it's a file or directory
	info, err := os.Stat(s.selectedPath)
	if err != nil {
		s.dragDropLabel.SetText("[ Drag & Drop your files here ]")
		s.fileInfoLabel.SetText("Error: " + err.Error())
		s.commentsEntry.SetText("")
		return
	}

	if info.IsDir() {
		s.dragDropLabel.SetText("📁 " + fileName)
		modeText := "Archive + Encrypt"
		if s.recursiveMode {
			modeText = "Recursive per-file encryption"
		}
		// Count files (non-recursive quick info)
		var fileCount int
		filepath.Walk(s.selectedPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil { return nil }
			if fi.IsDir() { return nil }
			lower := strings.ToLower(path)
			if strings.HasSuffix(lower, ".hadescrypt") || strings.HasSuffix(lower, ".heistcrypt") || strings.HasSuffix(lower, ".gpg") { return nil }
			fileCount++
			return nil
		})
		s.fileInfoLabel.SetText(fmt.Sprintf("Folder: %d file(s) | Mode: %s", fileCount, modeText))
		// Don't clear comments for directories, user might want to add them
	} else {
		// Check sidecar meta indicating archived folder
		metaPath := s.selectedPath + ".meta"
		if _, err := os.Stat(metaPath); err == nil {
			// Show folder-archive icon and parse minimal info
			data, _ := os.ReadFile(metaPath)
			// crude parse (avoid adding JSON dep): look for file_count & original_folder
			fileCount := "?"
			original := fileName
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "\"file_count\"") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 { fileCount = strings.Trim(parts[1], " ,\"") }
				}
				if strings.HasPrefix(line, "\"original_folder\"") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 { original = strings.Trim(parts[1], " ,\"") }
				}
			}
			s.dragDropLabel.SetText("📦 " + original)
			s.fileInfoLabel.SetText(fmt.Sprintf("Archived Folder (files: %s)", fileCount))
			return
		}
		s.dragDropLabel.SetText("📄 " + fileName)
		sizeText := uiutil.HumanBytes(info.Size())
		
		// Try to get detailed file information for encrypted files
		fileInfo, err := cryptoengine.GetFileInfo(s.selectedPath)
		if err == nil {
			format := fileInfo["format"].(string)
			
			if format == "HadesCrypt" {
				// HadesCrypt encrypted file
				modeName := fileInfo["encryption_mode_name"].(string)
				s.fileInfoLabel.SetText(fmt.Sprintf("🔒 Size: %s - %s", sizeText, modeName))
				
				// Extract and display comments
				if comments, ok := fileInfo["comments"].(string); ok && comments != "" {
					s.commentsEntry.SetText(comments)
					s.comments = comments
				} else {
					// Clear comments for encrypted files with no comments
					s.commentsEntry.SetText("")
					s.comments = ""
				}
			} else if format == "GnuPG/OpenPGP" {
				// GnuPG encrypted file
				s.fileInfoLabel.SetText(fmt.Sprintf("🔐 Size: %s - GnuPG/OpenPGP", sizeText))
				// GnuPG files don't store comments, but don't clear existing ones
			} else {
				// Regular file
				s.fileInfoLabel.SetText(fmt.Sprintf("Size: %s", sizeText))
				// Don't clear comments for regular files
			}
		} else {
			// Fallback for files that can't be analyzed
			s.fileInfoLabel.SetText(fmt.Sprintf("Size: %s", sizeText))
		}
	}
}

func (s *AppState) updateStrength(password string) {
	score, label := pw.StrengthScore(password)
	s.strengthBar.SetValue(score)
	s.strengthLabel.SetText("Strength: " + label)
}

func (s *AppState) validatePasswordMatch() {
	if s.password == "" && s.confirmPassword == "" {
		s.passwordMatchLabel.SetText("")
		return
	}
	
	if s.confirmPassword == "" {
		s.passwordMatchLabel.SetText("")
		return
	}
	
	if s.password == s.confirmPassword {
		// Passwords match - show green checkmark with animation
		s.passwordMatchLabel.SetText("✅ Match")
		
		// Animate the confirmation (simple color/text effect)
		go s.animatePasswordMatch(true)
	} else {
		// Passwords don't match - show red X
		s.passwordMatchLabel.SetText("❌ No Match")
		
		// Animate the mismatch
		go s.animatePasswordMatch(false)
	}
}

func (s *AppState) animatePasswordMatch(isMatch bool) {
	if isMatch {
		// Success animation - pulse green checkmark
		for i := 0; i < 3; i++ {
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() {
				s.passwordMatchLabel.SetText("✨ Match")
			})
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() {
				s.passwordMatchLabel.SetText("✅ Match")
			})
		}
	} else {
		// Error animation - pulse red X
		for i := 0; i < 2; i++ {
			time.Sleep(150 * time.Millisecond)
			fyne.Do(func() {
				s.passwordMatchLabel.SetText("⚠️ No Match")
			})
			time.Sleep(150 * time.Millisecond)
			fyne.Do(func() {
				s.passwordMatchLabel.SetText("❌ No Match")
			})
		}
	}
}

func (s *AppState) doEncrypt(w fyne.Window) {
	s.cancelRequested.Store(false)
	if s.selectedPath == "" && len(s.selectedPaths) == 0 {
		dialog.ShowInformation("Select input", "Please select a file, folder, or multiple files to encrypt.", w)
		return
	}
	if s.password == "" {
		dialog.ShowInformation("Password required", "Please enter a password.", w)
		return
	}
	if s.password != s.confirmPassword {
		dialog.ShowInformation("Password Mismatch", "Password and confirmation password do not match.", w)
		return
	}

	var singleInfo os.FileInfo
	var outputPath string
	if s.selectedPath != "" {
		var err error
		outputPath = s.defaultOutputPathForEncrypt(s.selectedPath)
		singleInfo, err = os.Stat(s.selectedPath)
		if err != nil { dialog.ShowError(err, w); return }
		if singleInfo.IsDir() && !s.recursiveMode { /* archive mode comment */ }
	}

	s.statusLabel.SetText("Encrypting…")
	s.progressBar.SetValue(0)

    go func() {
		onProgress := func(done, total int64) {
			fyne.Do(func() {
				if s.cancelRequested.Load() { return }
				if total <= 0 {
					s.progressBar.SetValue(0)
                return
            }
				s.progressBar.SetValue(float64(done) / float64(total))
			})
        }

        start := time.Now()
		var encErr error
		finalPassword := []byte(s.password)
		if s.keyfileManager.HasKeyfiles() { finalPassword = s.keyfileManager.GetCombinedKey([]byte(s.password)) }

		if len(s.selectedPaths) > 0 { // multi-file mode
			// Aggregate bytes across files & folders
			grandTotal, _ := s.computeMixedSelectionSize(s.recursiveMode)
			var processed int64
			for idx, p := range s.selectedPaths {
				if s.cancelRequested.Load() { encErr = fmt.Errorf("canceled"); break }
				fi, err := os.Stat(p); if err != nil { continue }
				base := filepath.Base(p)
				fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("Encrypting %d/%d: %s", idx+1, len(s.selectedPaths), base)) })
				if fi.IsDir() {
					// Choose strategy: recursive or archive
					if s.recursiveMode {
						cerr := s.encryptDirectoryRecursive(p, finalPassword, func(done,total int64){ if grandTotal>0 { onProgress(processed+done, grandTotal) } })
						if cerr != nil { encErr = cerr; break }
						// after folder, increment processed by summed size of its contents
						filepath.Walk(p, func(sp string, info os.FileInfo, e error) error {
							if e!=nil || info==nil || info.IsDir() { return nil }
							low := strings.ToLower(sp)
							if strings.HasSuffix(low, ".hadescrypt") || strings.HasSuffix(low, ".heistcrypt") || strings.HasSuffix(low, ".gpg") || strings.HasSuffix(low, ".pgp") { return nil }
							processed += info.Size(); return nil
						})
					} else {
						outArchive := s.defaultOutputPathForEncrypt(p)
						cerr := s.encryptDirectory(p, outArchive, finalPassword, func(done,total int64){ if grandTotal>0 { onProgress(processed+done/2, grandTotal) } })
						if cerr != nil { encErr = cerr; break }
						// After archive encryption, approximate processed as full folder content size
						filepath.Walk(p, func(sp string, info os.FileInfo, e error) error {
							if e!=nil || info==nil || info.IsDir() { return nil }
							low := strings.ToLower(sp)
							if strings.HasSuffix(low, ".hadescrypt") || strings.HasSuffix(low, ".heistcrypt") || strings.HasSuffix(low, ".gpg") || strings.HasSuffix(low, ".pgp") { return nil }
							processed += info.Size(); return nil
						})
					}
					// history entry folder
					s.config.AddHistoryEntry(config.HistoryEntry{FileName: base, Operation:"encrypt-folder", Size: fi.Size(), Timestamp: time.Now().Unix(), Result: "success"})
					if s.deleteAfter { os.RemoveAll(p) }
				} else if fi.Mode().IsRegular() {
					out := s.defaultOutputPathForEncrypt(p)
					cerr := cryptoengine.EncryptFileWithMode(p, out, finalPassword, s.encryptionMode, func(done,total int64){ if grandTotal>0 { onProgress(processed+done, grandTotal) } })
					if cerr != nil { encErr = cerr; break }
					processed += fi.Size()
					s.config.AddHistoryEntry(config.HistoryEntry{FileName: base, Operation:"encrypt", Size: fi.Size(), Timestamp: time.Now().Unix(), Result: "success"})
					if s.deleteAfter { os.Remove(p) }
				}
				if onProgress != nil { onProgress(processed, grandTotal) }
			}
			elapsed := time.Since(start).Round(time.Millisecond)
			if encErr == nil { fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("✅ Encrypted %d item(s) in %s", len(s.selectedPaths), elapsed)) }) }
		} else if singleInfo != nil && singleInfo.IsDir() {
			if s.recursiveMode { encErr = s.encryptDirectoryRecursive(s.selectedPath, finalPassword, onProgress) } else { encErr = s.encryptDirectory(s.selectedPath, outputPath, finalPassword, onProgress) }
			elapsed := time.Since(start).Round(time.Millisecond)
			if encErr == nil { fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("✅ Encrypted folder in %s", elapsed)) }) }
		} else {
			encErr = cryptoengine.EncryptFileWithMode(s.selectedPath, outputPath, finalPassword, s.encryptionMode, onProgress)
			elapsed := time.Since(start).Round(time.Millisecond)
			if encErr == nil { fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("✅ Encrypted %s in %s", filepath.Base(s.selectedPath), elapsed)) }) }
			// single file history
			s.config.AddHistoryEntry(config.HistoryEntry{FileName: filepath.Base(s.selectedPath), Operation:"encrypt", Size: singleInfo.Size(), Timestamp: time.Now().Unix(), Result: "success"})
			if s.deleteAfter { os.Remove(s.selectedPath) }
		}

		// Save config/history at end
		s.config.Save()

		// (legacy per-file final status removed; handled inline per branch)
	}()
}

func (s *AppState) doDecrypt(w fyne.Window) {
	s.cancelRequested.Store(false)
	if s.selectedPath == "" && len(s.selectedPaths) == 0 {
		dialog.ShowInformation("Select input", "Please select a file, folder, or multiple encrypted items to decrypt.", w)
		return
	}
	if s.password == "" {
        dialog.ShowInformation("Password required", "Please enter a password.", w)
        return
    }

	outputPath := ""
	if s.selectedPath != "" { outputPath = s.defaultOutputPathForDecrypt(s.selectedPath) }

	s.statusLabel.SetText("Decrypting…")
	s.progressBar.SetValue(0)

	go func() {
		// Batch multi-selection path
		if len(s.selectedPaths) > 0 {
			finalPassword := []byte(s.password)
			if s.keyfileManager.HasKeyfiles() { finalPassword = s.keyfileManager.GetCombinedKey([]byte(s.password)) }
			// Collect targets (files or directories)
			var targets []string
			for _, p := range s.selectedPaths { targets = append(targets, p) }
			// Pre-compute total bytes (approx): for encrypted dirs (user selected) we'll walk inside
			var totalBytes int64
			for _, t := range targets {
				fi, err := os.Stat(t); if err != nil { continue }
				if fi.IsDir() {
					filepath.Walk(t, func(sp string, info os.FileInfo, e error) error {
						if e!=nil || info==nil || info.IsDir() { return nil }
						low:=strings.ToLower(sp)
						if strings.HasSuffix(low, ".hadescrypt") || strings.HasSuffix(low, ".heistcrypt") || strings.HasSuffix(low, ".gpg") || strings.HasSuffix(low, ".pgp") { totalBytes += info.Size() }
						return nil
					})
				} else if fi.Mode().IsRegular() { totalBytes += fi.Size() }
			}
			var processed int64
			start := time.Now()
			for idx, t := range targets {
				if s.cancelRequested.Load() { break }
				fi, err := os.Stat(t); if err != nil { continue }
				base := filepath.Base(t)
				fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("Decrypting %d/%d: %s", idx+1, len(targets), base)) })
				if fi.IsDir() {
					// Decrypt all encrypted files inside directory recursively
					dErr := s.decryptDirectoryRecursive(t, finalPassword, func(done,total int64){ if totalBytes>0 { fyne.Do(func(){ s.progressBar.SetValue(float64(processed+done)/float64(totalBytes)) }) } })
					// After finishing dir, increment processed by sizes of encrypted files within
					filepath.Walk(t, func(sp string, info os.FileInfo, e error) error {
						if e!=nil || info==nil || info.IsDir() { return nil }
						low:=strings.ToLower(sp)
						if strings.HasSuffix(low, ".hadescrypt") || strings.HasSuffix(low, ".heistcrypt") || strings.HasSuffix(low, ".gpg") || strings.HasSuffix(low, ".pgp") { processed += info.Size() }
						return nil })
					if dErr != nil { fyne.Do(func(){ s.statusLabel.SetText("Error: "+dErr.Error()) }); break }
				} else {
					out := s.defaultOutputPathForDecrypt(t)
					var dErr error
					if s.isHadesCryptFile(t) { dErr = s.decryptFileAuto(t, out, finalPassword, func(done,total int64){ if totalBytes>0 { fyne.Do(func(){ s.progressBar.SetValue(float64(processed+done)/float64(totalBytes)) }) } })
					} else if s.isGnuPGFile(t) { dErr = cryptoengine.DecryptFileWithGnuPG(t, out, finalPassword, func(done,total int64){ if totalBytes>0 { fyne.Do(func(){ s.progressBar.SetValue(float64(processed+done)/float64(totalBytes)) }) } })
					} else { dErr = cryptoengine.DecryptFile(t, out, finalPassword, s.forceDecrypt, func(done,total int64){ if totalBytes>0 { fyne.Do(func(){ s.progressBar.SetValue(float64(processed+done)/float64(totalBytes)) }) } }) }
					if dErr != nil { fyne.Do(func(){ s.statusLabel.SetText("Error: "+dErr.Error()) }); break }
					processed += fi.Size()
				}
				fyne.Do(func(){ if totalBytes>0 { s.progressBar.SetValue(float64(processed)/float64(totalBytes)) } })
				if s.deleteAfter { os.RemoveAll(t) }
			}
			if !s.cancelRequested.Load() {
				elapsed := time.Since(start).Round(time.Millisecond)
				fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("✅ Decrypted %d item(s) in %s", len(targets), elapsed)) })
			} else { fyne.Do(func(){ s.statusLabel.SetText("Canceled") }) }
			return
		}
		// If single selectedPath is a directory: decrypt all encrypted files inside.
		if s.selectedPath != "" { if info, err := os.Stat(s.selectedPath); err == nil && info.IsDir() {
			start := time.Now()
			finalPassword := []byte(s.password)
			if s.keyfileManager.HasKeyfiles() { finalPassword = s.keyfileManager.GetCombinedKey([]byte(s.password)) }
			err := s.decryptDirectoryRecursive(s.selectedPath, finalPassword, func(done,total int64){ fyne.Do(func(){ if total>0 { s.progressBar.SetValue(float64(done)/float64(total)) } }) })
			elapsed := time.Since(start).Round(time.Millisecond)
			fyne.Do(func(){
				if err != nil { s.statusLabel.SetText("Error: "+err.Error()) } else { s.statusLabel.SetText(fmt.Sprintf("✅ Folder decrypted in %s", elapsed)) }
			})
			return
		} }
		onProgress := func(done, total int64) {
			fyne.Do(func() {
				if total <= 0 {
					s.progressBar.SetValue(0)
                return
            }
				s.progressBar.SetValue(float64(done) / float64(total))
			})
        }

        start := time.Now()
		
		// Prepare password with keyfiles
		finalPassword := []byte(s.password)
		if s.keyfileManager.HasKeyfiles() {
			finalPassword = s.keyfileManager.GetCombinedKey([]byte(s.password))
		}

		// Auto decrypt for HadesCrypt (.hadescrypt/.heistcrypt) – handles single-file or archived folder transparently
		var err error
		if s.isHadesCryptFile(s.selectedPath) {
			err = s.decryptFileAuto(s.selectedPath, outputPath, finalPassword, onProgress)
		} else {
			if s.isGnuPGFile(s.selectedPath) {
				err = cryptoengine.DecryptFileWithGnuPG(s.selectedPath, outputPath, finalPassword, onProgress)
			} else {
				err = cryptoengine.DecryptFile(s.selectedPath, outputPath, finalPassword, s.forceDecrypt, onProgress)
			}
		}
		
		elapsed := time.Since(start).Round(time.Millisecond)
		
		// Get file size for history
		var fileSize int64
		if info, statErr := os.Stat(s.selectedPath); statErr == nil {
			fileSize = info.Size()
		}

		// Add to history
		historyEntry := config.HistoryEntry{
			FileName:  filepath.Base(s.selectedPath),
			Operation: "decrypt",
			Size:      fileSize,
			Timestamp: time.Now().Unix(),
		}

		fyne.Do(func() {
			if err != nil {
				if strings.Contains(err.Error(), "canceled") {
					s.statusLabel.SetText("Canceled")
					return
				}
				historyEntry.Result = "error"
				historyEntry.Error = err.Error()
				s.statusLabel.SetText("Error: " + err.Error())
            dialog.ShowError(err, w)
			} else {
				historyEntry.Result = "success"
				statusMsg := fmt.Sprintf("✅ Decrypted to %s in %s", filepath.Base(outputPath), elapsed)
				
				// Delete source file if option is enabled
				if s.deleteAfter {
					if deleteErr := os.Remove(s.selectedPath); deleteErr != nil {
						statusMsg += " (Warning: Could not delete source)"
					} else {
						statusMsg += " (Source deleted)"
					}
				}
				
				s.statusLabel.SetText(statusMsg)
			}
		})

		s.config.AddHistoryEntry(historyEntry)
		s.config.Save()
	}()
}

func (s *AppState) encryptDirectory(inputDir, outputPath string, password []byte, onProgress cryptoengine.ProgressCallback) error {
	// Create temporary tar.gz file
	tempArchive := outputPath + ".temp.tar.gz"
	defer os.Remove(tempArchive)

	// Phase 1: create archive (0-50%)
	err := archiver.CreateTarGz(inputDir, tempArchive, func(processed, total int64) {
		if onProgress != nil && total > 0 {
			progress := float64(processed) / float64(total) * 0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}

	// Compute SHA-256 of plaintext archive for integrity metadata
	archiveHash := ""
	if f, herr := os.Open(tempArchive); herr == nil {
		func() {
			defer f.Close()
			h := sha256.New()
			io.Copy(h, f)
			archiveHash = hex.EncodeToString(h.Sum(nil))
		}()
	}

	// Phase 2: encrypt archive (50-100%)
	err = cryptoengine.EncryptFile(tempArchive, outputPath, password, func(processed, total int64) {
		if onProgress != nil && total > 0 {
			progress := 0.5 + (float64(processed)/float64(total))*0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("encrypt archive: %w", err)
	}

	// Sidecar metadata (.meta JSON)
	metaPath := outputPath + ".meta"
	var fileCount int
	var totalBytes int64
	filepath.Walk(inputDir, func(p string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info != nil && !info.IsDir() { fileCount++; totalBytes += info.Size() }
		return nil
	})
	metaJSON := fmt.Sprintf("{\n  \"type\": \"archive-folder\",\n  \"original_folder\": %q,\n  \"file_count\": %d,\n  \"total_size\": %d,\n  \"archive_sha256\": %q\n}", filepath.Base(inputDir), fileCount, totalBytes, archiveHash)
	os.WriteFile(metaPath, []byte(metaJSON), 0600)

	return nil
}

// encryptDirectoryRecursive walks a directory and encrypts each file individually preserving structure.
// Each file produces <name>.hadescrypt (or .gpg) beside original. Progress aggregated by total bytes.
func (s *AppState) encryptDirectoryRecursive(inputDir string, password []byte, onProgress cryptoengine.ProgressCallback) error {
	var totalBytes int64
	var files []string
	// Collect files
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() { return nil }
		// Skip already encrypted outputs
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".hadescrypt") || strings.HasSuffix(lower, ".heistcrypt") || strings.HasSuffix(lower, ".gpg") { return nil }
		files = append(files, path)
		totalBytes += info.Size()
		return nil
	})
	if err != nil { return err }
	if totalBytes == 0 { return fmt.Errorf("no files to encrypt in directory") }

	var processedBytes int64
	for _, file := range files {
		if s.cancelRequested.Load() { return fmt.Errorf("canceled") }
		rel, _ := filepath.Rel(inputDir, file)
		// progress callback for single file
		fi, _ := os.Stat(file)
		singleSize := fi.Size()
		fileOutput := file + ".hadescrypt"
		if s.encryptionMode == cryptoengine.ModeGnuPG { fileOutput = file + ".gpg" }
		err := cryptoengine.EncryptFileWithMode(file, fileOutput, password, s.encryptionMode, func(done, total int64){
			// translate per-file progress into global progress (estimate): processedBytes + done
			if onProgress != nil && totalBytes > 0 {
				onProgress(processedBytes+done, totalBytes)
			}
		})
		if err != nil { return fmt.Errorf("encrypt %s: %w", rel, err) }
		processedBytes += singleSize
		if onProgress != nil { onProgress(processedBytes, totalBytes) }
		if s.deleteAfter { os.Remove(file) }
	}
	return nil
}

func (s *AppState) decryptDirectory(encryptedFile, outputDir string, password []byte, onProgress cryptoengine.ProgressCallback) error {
	// Create temporary file for decrypted archive
	tempArchive := encryptedFile + ".temp.tar.gz"
	defer os.Remove(tempArchive) // Clean up temp file

	// First decrypt the file
	err := cryptoengine.DecryptFile(encryptedFile, tempArchive, password, false, func(processed, total int64) {
		// Report progress for decryption phase (0-50%)
		if onProgress != nil && total > 0 {
			progress := float64(processed) / float64(total) * 0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("decrypt file: %w", err)
	}

	// Check if the decrypted file is actually a tar.gz archive
	if !archiver.IsArchive(tempArchive) {
		// If it's not an archive, just rename it to the output path
		return os.Rename(tempArchive, outputDir)
	}

	// Extract the archive
	err = archiver.ExtractTarGz(tempArchive, outputDir, func(processed, total int64) {
		// Report progress for extraction phase (50-100%)
		if onProgress != nil && total > 0 {
			progress := 0.5 + (float64(processed)/float64(total))*0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	return nil
}

// decryptFileAuto decrypts a hades/heist crypt file then inspects if decrypted result is a gzip tar archive.
// If archive: extracts into a directory (outputPath) and removes temp decrypted file.
// If not archive: keeps decrypted file.
func (s *AppState) decryptFileAuto(encryptedFile, outputPath string, password []byte, onProgress cryptoengine.ProgressCallback) error {
	// Read header quickly for integrity (HadesCrypt only)
	var expectedSize int64 = -1
	if s.isHadesCryptFile(encryptedFile) {
		f, err := os.Open(encryptedFile)
		if err == nil {
			defer f.Close()
			buf := make([]byte, 4)
			if _, err := io.ReadFull(f, buf); err == nil && string(buf) == "HAD1" {
				// version
				ver := make([]byte,1); io.ReadFull(f,ver)
				mode := make([]byte,1); io.ReadFull(f,mode)
				salt := make([]byte,16); io.ReadFull(f,salt)
				nonce := make([]byte,8); io.ReadFull(f,nonce)
				cs := make([]byte,4); io.ReadFull(f,cs)
				osz := make([]byte,8); if _, err := io.ReadFull(f,osz); err==nil { expectedSize = int64(binary.BigEndian.Uint64(osz)) }
			}
		}
	}
	tempDecrypted := encryptedFile + ".__dec_tmp__"
	defer os.Remove(tempDecrypted)
	// low-level decrypt (not directory)
	err := cryptoengine.DecryptFile(encryptedFile, tempDecrypted, password, s.forceDecrypt, onProgress)
	if err != nil { return err }
	// Check if decrypted is archive
	if archiver.IsArchive(tempDecrypted) {
		// Optional hash verification via sidecar meta
		metaPath := encryptedFile + ".meta"
		if data, rerr := os.ReadFile(metaPath); rerr == nil {
			// crude parse for archive_sha256
			lines := strings.Split(string(data), "\n")
			var expectedHash string
			for _, ln := range lines {
				if strings.Contains(ln, "archive_sha256") {
					parts := strings.Split(ln, ":")
					if len(parts) >= 2 {
						v := strings.TrimSpace(parts[1])
						v = strings.Trim(v, ",")
						v = strings.Trim(v, "\"")
						expectedHash = v
						break
					}
				}
			}
			if expectedHash != "" {
				f, herr := os.Open(tempDecrypted)
				if herr == nil {
					h := sha256.New()
					io.Copy(h, f)
					f.Close()
					calc := hex.EncodeToString(h.Sum(nil))
					if !strings.EqualFold(calc, expectedHash) {
						fyne.Do(func(){ s.statusLabel.SetText("❌ Hash mismatch — decryption aborted") })
						return fmt.Errorf("archive hash mismatch (expected %s got %s)", expectedHash, calc)
					}
					fyne.Do(func(){ s.statusLabel.SetText("🔐 Hash verified OK — extracting...") })
				}
			}
		}
		// Ensure directory target
		if err := os.MkdirAll(outputPath, 0755); err != nil { return err }
		var archCb archiver.ProgressCallback
		if onProgress != nil { archCb = func(done,total int64){ onProgress(done,total) } }
		if err := archiver.ExtractTarGz(tempDecrypted, outputPath, archCb); err != nil {
			return fmt.Errorf("extract archive: %w", err)
		}
		// Remove sidecar meta if exists
		os.Remove(metaPath)
		return nil
	}
	// Not archive -> move/rename to outputPath
	if err := os.Rename(tempDecrypted, outputPath); err != nil { return err }
	if expectedSize >= 0 {
		if fi, err := os.Stat(outputPath); err == nil && fi.Size() != expectedSize {
			return fmt.Errorf("integrity warning: size mismatch expected %d got %d", expectedSize, fi.Size())
		}
	}
	return nil
}

// decryptDirectoryRecursive decrypts every encrypted file within a directory tree.
// It handles .hadescrypt, .heistcrypt, .gpg, .pgp files. Output overwrites by stripping extension.
func (s *AppState) decryptDirectoryRecursive(root string, password []byte, onProgress cryptoengine.ProgressCallback) error {
	var encryptedFiles []string
	var totalBytes int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.IsDir() { return nil }
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".hadescrypt") || strings.HasSuffix(lower, ".heistcrypt") || strings.HasSuffix(lower, ".gpg") || strings.HasSuffix(lower, ".pgp") {
			encryptedFiles = append(encryptedFiles, path)
			totalBytes += info.Size()
		}
		return nil
	})
	if err != nil { return err }
	if len(encryptedFiles) == 0 { return fmt.Errorf("no encrypted files found in folder") }

	var processedBytes int64
	for i, file := range encryptedFiles {
		if s.cancelRequested.Load() { return fmt.Errorf("canceled") }
		rel, _ := filepath.Rel(root, file)
		fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("Decrypting %d/%d: %s", i+1, len(encryptedFiles), rel)) })
		outPath := s.defaultOutputPathForDecrypt(file)
		fi, _ := os.Stat(file)
		size := fi.Size()
		// choose method
		var derr error
		if s.isGnuPGFile(file) {
			derr = cryptoengine.DecryptFileWithGnuPG(file, outPath, password, func(done,total int64){ if onProgress!=nil { onProgress(processedBytes+done,totalBytes) } })
		} else if s.isHadesCryptFile(file) {
			derr = s.decryptFileAuto(file, outPath, password, func(done,total int64){ if onProgress!=nil { onProgress(processedBytes+done,totalBytes) } })
		} else {
			derr = cryptoengine.DecryptFile(file, outPath, password, s.forceDecrypt, func(done,total int64){ if onProgress!=nil { onProgress(processedBytes+done,totalBytes) } })
		}
		if derr != nil { return fmt.Errorf("decrypt %s: %w", rel, derr) }
		// history entry
		hist := config.HistoryEntry{FileName: rel, Operation: "decrypt", Size: size, Timestamp: time.Now().Unix(), Result: "success"}
		s.config.AddHistoryEntry(hist)
		processedBytes += size
		if onProgress != nil { onProgress(processedBytes, totalBytes) }
		if s.deleteAfter { os.Remove(file) }
	}
	fyne.Do(func(){ s.statusLabel.SetText(fmt.Sprintf("✅ Decrypted %d files", len(encryptedFiles))) })
	s.config.Save()
	return nil
}

func (s *AppState) defaultOutputPathForEncrypt(inPath string) string {
	// Use appropriate extension based on encryption mode
	if s.encryptionMode == cryptoengine.ModeGnuPG {
		return inPath + ".gpg"
	}
	return inPath + ".hadescrypt"
}

func (s *AppState) defaultOutputPathForDecrypt(inPath string) string {
	lowerPath := strings.ToLower(inPath)
	
	// Handle GnuPG files
	if strings.HasSuffix(lowerPath, ".gpg") {
		return strings.TrimSuffix(inPath, ".gpg")
	}
	if strings.HasSuffix(lowerPath, ".pgp") {
		return strings.TrimSuffix(inPath, ".pgp")
	}
	
    // Handle HadesCrypt files
    if strings.HasSuffix(lowerPath, ".hades") {
        return strings.TrimSuffix(inPath, ".hades")
    }
    if strings.HasSuffix(lowerPath, ".hadescrypt") || strings.HasSuffix(lowerPath, ".heistcrypt") {
        return strings.TrimSuffix(inPath, filepath.Ext(inPath))
    }
	
    return inPath + ".dec"
}

// isGnuPGFile checks if the file is a GnuPG/OpenPGP file
func (s *AppState) isGnuPGFile(filePath string) bool {
	// Check by extension first
	lowerPath := strings.ToLower(filePath)
	if strings.HasSuffix(lowerPath, ".gpg") || strings.HasSuffix(lowerPath, ".pgp") {
		return true
	}
	
	// Check by file content (OpenPGP magic bytes)
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Read first few bytes to check for OpenPGP format
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return false
	}
	
	// OpenPGP files typically start with specific packet headers
	if len(header) > 0 {
		firstByte := header[0]
		// Check for OpenPGP packet format (high bit set indicates OpenPGP packet)
		if (firstByte&0x80) != 0 {
			return true
		}
	}
	
	return false
}

// isHadesCryptFile detects files produced by HadesCrypt (.hadescrypt or .heistcrypt) using extension and magic header.
func (s *AppState) isHadesCryptFile(path string) bool {
	lower := strings.ToLower(path)
	if !(strings.HasSuffix(lower, ".hadescrypt") || strings.HasSuffix(lower, ".heistcrypt")) {
		return false
	}
	f, err := os.Open(path)
	if err != nil { return false }
	defer f.Close()
	header := make([]byte, 4)
	if _, err := f.Read(header); err != nil { return false }
	return string(header) == "HAD1"
}

func (s *AppState) updateKeyfilesDisplay() {
	count := s.keyfileManager.Count()
	if count == 0 {
		s.keyfilesLabel.SetText("No keyfiles selected")
	} else {
		s.keyfilesLabel.SetText(fmt.Sprintf("%d keyfile(s) selected", count))
	}
	s.keyfilesList.Refresh()
}

func (s *AppState) showKeyfileDialog(w fyne.Window) {
	fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if rc == nil {
			return
		}
		defer rc.Close()
		
		path := rc.URI().Path()
		
		// Validate keyfile
		if err := keyfiles.ValidateKeyfile(path); err != nil {
			dialog.ShowError(fmt.Errorf("Invalid keyfile: %v", err), w)
        return
    }
		
		// Add keyfile
		if err := s.keyfileManager.AddKeyfile(path); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to add keyfile: %v", err), w)
        return
    }
		
		s.updateKeyfilesDisplay()
	}, w)
	
	fd.SetFilter(nil) // Allow all files
	fd.Show()
}

func (s *AppState) showGenerateKeyfileDialog(w fyne.Window) {
	sizeEntry := widget.NewEntry()
	sizeEntry.SetText("1")
	
	unitSelect := widget.NewSelect([]string{"KB", "MB"}, func(string) {})
	unitSelect.SetSelected("KB")
	
	sizeRow := container.NewHBox(
		widget.NewLabel("Size:"),
		sizeEntry,
		unitSelect,
	)
	
	content := container.NewVBox(
		widget.NewLabel("Generate a secure keyfile"),
		sizeRow,
	)
	
	d := dialog.NewCustomConfirm("Generate Keyfile", "Generate", "Cancel", content, func(generate bool) {
		if !generate {
			return
		}
		
		// Parse size
		sizeStr := sizeEntry.Text
		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			dialog.ShowError(fmt.Errorf("Invalid size: %s", sizeStr), w)
                return
            }
		
		// Convert to KB
		if unitSelect.Selected == "MB" {
			size *= 1024
        }
		
		// Show save dialog
		fd := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
			if uc == nil {
				return
			}
			
			outputPath := uc.URI().Path()
			uc.Close()
			
			// Generate keyfile
			if err := keyfiles.GenerateKeyfile(outputPath, size); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to generate keyfile: %v", err), w)
        return
    }
			
			// Add to manager
			if err := s.keyfileManager.AddKeyfile(outputPath); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to add generated keyfile: %v", err), w)
        return
    }
			
			s.updateKeyfilesDisplay()
			dialog.ShowInformation("Success", fmt.Sprintf("Keyfile generated and added: %s", filepath.Base(outputPath)), w)
		}, w)
		
		fd.SetFileName("keyfile.key")
		fd.Show()
	}, w)
	
	d.Show()
}

func (s *AppState) showPasswordGeneratorDialog(w fyne.Window) {
	// Length slider
	lengthSlider := widget.NewSlider(8, 128)
	lengthSlider.SetValue(20)
	lengthLabel := widget.NewLabel("20")
	lengthSlider.OnChanged = func(value float64) {
		lengthLabel.SetText(fmt.Sprintf("%.0f", value))
	}
	
	// Character type checkboxes
	lowerCheck := widget.NewCheck("Lowercase (a-z)", nil)
	lowerCheck.SetChecked(true)
	upperCheck := widget.NewCheck("Uppercase (A-Z)", nil)
	upperCheck.SetChecked(true)
	digitsCheck := widget.NewCheck("Digits (0-9)", nil)
	digitsCheck.SetChecked(true)
	symbolsCheck := widget.NewCheck("Symbols (!@#$...)", nil)
	symbolsCheck.SetChecked(true)
	
	// Preview
	previewEntry := widget.NewEntry()
	previewEntry.SetText("")
	
	// Generate function
	generatePassword := func() {
		opts := pw.GenerateOptions{
			Length:     int(lengthSlider.Value),
			UseLower:   lowerCheck.Checked,
			UseUpper:   upperCheck.Checked,
			UseDigits:  digitsCheck.Checked,
			UseSymbols: symbolsCheck.Checked,
		}
		
		password, err := pw.GenerateWithOptions(opts)
        if err != nil {
            dialog.ShowError(err, w)
            return
        }
		previewEntry.SetText(password)
	}
	
	// Generate initial password
	generatePassword()
	
	// Auto-generate on option change
	lowerCheck.OnChanged = func(bool) { generatePassword() }
	upperCheck.OnChanged = func(bool) { generatePassword() }
	digitsCheck.OnChanged = func(bool) { generatePassword() }
	symbolsCheck.OnChanged = func(bool) { generatePassword() }
	lengthSlider.OnChanged = func(float64) { generatePassword() }
	
	// Regenerate button
	regenBtn := widget.NewButton("Regenerate", func() {
		generatePassword()
	})
	
	content := container.NewVBox(
		widget.NewLabel("Password Generator"),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Length:"), lengthSlider, lengthLabel),
		widget.NewSeparator(),
		lowerCheck,
		upperCheck,
		digitsCheck,
		symbolsCheck,
		widget.NewSeparator(),
		widget.NewLabel("Preview:"),
		previewEntry,
		regenBtn,
	)
	
	d := dialog.NewCustomConfirm("Generate Password", "Use Password", "Cancel", content, func(use bool) {
		if use && previewEntry.Text != "" {
			s.passwordEntry.SetText(previewEntry.Text)
			s.confirmPasswordEntry.SetText(previewEntry.Text)
			s.password = previewEntry.Text
			s.confirmPassword = previewEntry.Text
			s.updateStrength(previewEntry.Text)
			s.validatePasswordMatch()
		}
	}, w)
	
	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}

func (s *AppState) buildAdvancedPanel() *widget.Accordion {
	// Initialize defaults
	s.splitSize = 100
	s.splitUnit = "MiB"

	deleteCheck := widget.NewCheck("Delete source files after operation", func(checked bool) {
		s.deleteAfter = checked
	})
	deleteCheck.SetChecked(true) // Set as default
	
	keyfilesCheck := widget.NewCheck("Use Keyfiles", func(checked bool) {
		s.useKeyfiles = checked
	})
	
	requireOrderCheck := widget.NewCheck("Require correct keyfile order", func(checked bool) {
		s.keyfileManager.RequireOrder = checked
	})
	
	paranoidCheck := widget.NewCheck("Paranoid Mode (XChaCha20 + Serpent)", func(checked bool) {
		s.paranoidMode = checked
	})
	
	rsCheck := widget.NewCheck("Reed-Solomon ECC (error correction)", func(checked bool) {
		s.reedSolomon = checked
	})
	
	forceCheck := widget.NewCheck("Force Decrypt (ignore integrity errors)", func(checked bool) {
		s.forceDecrypt = checked
	})
	
	splitCheck := widget.NewCheck("Split into chunks", func(checked bool) {
		s.splitOutput = checked
	})
	
	// Split size controls
	splitSizeEntry := widget.NewEntry()
	splitSizeEntry.SetText("100")
	splitSizeEntry.OnChanged = func(text string) {
		if size, err := strconv.Atoi(text); err == nil && size > 0 {
			s.splitSize = size
		}
	}
	
	splitUnitSelect := widget.NewSelect([]string{"KiB", "MiB", "GiB", "TiB"}, func(unit string) {
		s.splitUnit = unit
	})
	splitUnitSelect.SetSelected("MiB")
	
	splitRow := container.NewHBox(
		widget.NewLabel("Size:"),
		splitSizeEntry,
		splitUnitSelect,
	)
	
	compressCheck := widget.NewCheck("Compress files (Deflate)", func(checked bool) {
		s.compressFiles = checked
	})
	
	denyCheck := widget.NewCheck("Deniability Mode (hide encryption)", func(checked bool) {
		s.deniabilityMode = checked
	})
	
	recursiveCheck := widget.NewCheck("Recursive Mode (process files individually)", func(checked bool) {
		s.recursiveMode = checked
	})

    content := container.NewVBox(
		deleteCheck,
		widget.NewSeparator(),
		keyfilesCheck,
		container.NewPadded(requireOrderCheck),
		paranoidCheck,
		rsCheck,
		forceCheck,
		widget.NewSeparator(),
		splitCheck,
		container.NewPadded(splitRow),
		compressCheck,
		denyCheck,
		recursiveCheck,
	)
	
	item := widget.NewAccordionItem("Advanced Options ▼", content)
    acc := widget.NewAccordion(item)
    return acc
}
