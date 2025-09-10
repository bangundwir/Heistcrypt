package main

import (
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

	"github.com/bangundwir/HadesCrypt/internal/archiver"
	"github.com/bangundwir/HadesCrypt/internal/config"
	"github.com/bangundwir/HadesCrypt/internal/cryptoengine"
	"github.com/bangundwir/HadesCrypt/internal/keyfiles"
	pw "github.com/bangundwir/HadesCrypt/internal/password"
	uiutil "github.com/bangundwir/HadesCrypt/internal/ui"
)

type AppState struct {
	selectedPath        string
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

func main() {
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

    w := application.NewWindow("HadesCrypt 🔱 — Lock your secrets, rule your data.")
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

	// Drag & Drop Area
	s.dragDropLabel = widget.NewLabelWithStyle("[ Drag & Drop your files here ]", fyne.TextAlignCenter, fyne.TextStyle{})
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

	// File selection button as fallback
    selectBtn := widget.NewButton("Select File", func() {
		s.showFileDialog(w)
    })

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
		container.NewPadded(selectBtn),
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
		if len(uris) > 0 {
			path := uris[0].Path()
			s.setSelectedFile(path)
		}
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

func (s *AppState) setSelectedFile(path string) {
	s.selectedPath = path
	s.updateFileInfo()
}

func (s *AppState) updateFileInfo() {
	if s.selectedPath == "" {
		s.dragDropLabel.SetText("[ Drag & Drop your files here ]")
		s.fileInfoLabel.SetText("")
		return
	}

	fileName := filepath.Base(s.selectedPath)
	
	// Check if it's a file or directory
	info, err := os.Stat(s.selectedPath)
	if err != nil {
		s.dragDropLabel.SetText("[ Drag & Drop your files here ]")
		s.fileInfoLabel.SetText("Error: " + err.Error())
		return
	}

	if info.IsDir() {
		s.dragDropLabel.SetText("📁 " + fileName)
		s.fileInfoLabel.SetText("Folder selected")
	} else {
		s.dragDropLabel.SetText("📄 " + fileName)
		sizeText := uiutil.HumanBytes(info.Size())
		s.fileInfoLabel.SetText(fmt.Sprintf("Size: %s", sizeText))
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
	if s.selectedPath == "" {
		dialog.ShowInformation("Select file", "Please select a file or folder to encrypt.", w)
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

	outputPath := s.defaultOutputPathForEncrypt(s.selectedPath)
	
	// Check if input is a directory
	info, err := os.Stat(s.selectedPath)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	if info.IsDir() && !s.recursiveMode {
		dialog.ShowInformation("Folder encryption", "Please enable 'Recursive Mode' in Advanced Options to encrypt folders.", w)
		return
	}

	s.statusLabel.SetText("Encrypting…")
	s.progressBar.SetValue(0)

    go func() {
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
		var encErr error

		// Prepare password with keyfiles
		finalPassword := []byte(s.password)
		if s.keyfileManager.HasKeyfiles() {
			finalPassword = s.keyfileManager.GetCombinedKey([]byte(s.password))
		}

		if info.IsDir() {
			encErr = s.encryptDirectory(s.selectedPath, outputPath, finalPassword, onProgress)
		} else {
			encErr = cryptoengine.EncryptFileWithMode(s.selectedPath, outputPath, finalPassword, s.encryptionMode, onProgress)
		}

        elapsed := time.Since(start).Round(time.Millisecond)
		
		// Add to history
		historyEntry := config.HistoryEntry{
			FileName:  filepath.Base(s.selectedPath),
			Operation: "encrypt",
			Size:      info.Size(),
			Timestamp: time.Now().Unix(),
		}

		fyne.Do(func() {
			if encErr != nil {
				historyEntry.Result = "error"
				historyEntry.Error = encErr.Error()
				s.statusLabel.SetText("Error: " + encErr.Error())
				dialog.ShowError(encErr, w)
			} else {
				historyEntry.Result = "success"
				statusMsg := fmt.Sprintf("✅ Encrypted to %s in %s", filepath.Base(outputPath), elapsed)
				
				// Delete source file if option is enabled
				if s.deleteAfter {
					if deleteErr := os.RemoveAll(s.selectedPath); deleteErr != nil {
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

func (s *AppState) doDecrypt(w fyne.Window) {
	if s.selectedPath == "" {
        dialog.ShowInformation("Select file", "Please select a file to decrypt.", w)
        return
    }
	if s.password == "" {
        dialog.ShowInformation("Password required", "Please enter a password.", w)
        return
    }

	outputPath := s.defaultOutputPathForDecrypt(s.selectedPath)

	s.statusLabel.SetText("Decrypting…")
	s.progressBar.SetValue(0)

    go func() {
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

		// Check if this might be an encrypted folder (has .hadescrypt extension and might contain an archive)
		var err error
		isEncryptedFolder := strings.HasSuffix(strings.ToLower(s.selectedPath), ".hadescrypt")
		
		if isEncryptedFolder && s.recursiveMode {
			err = s.decryptDirectory(s.selectedPath, outputPath, finalPassword, onProgress)
		} else {
			err = cryptoengine.DecryptFile(s.selectedPath, outputPath, finalPassword, s.forceDecrypt, onProgress)
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
	defer os.Remove(tempArchive) // Clean up temp file

	// Create tar.gz archive
	err := archiver.CreateTarGz(inputDir, tempArchive, func(processed, total int64) {
		// Report progress for archiving phase (0-50%)
		if onProgress != nil && total > 0 {
			progress := float64(processed) / float64(total) * 0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}

	// Encrypt the archive
	err = cryptoengine.EncryptFile(tempArchive, outputPath, password, func(processed, total int64) {
		// Report progress for encryption phase (50-100%)
		if onProgress != nil && total > 0 {
			progress := 0.5 + (float64(processed)/float64(total))*0.5
			onProgress(int64(progress*float64(total)), total)
		}
	})
	if err != nil {
		return fmt.Errorf("encrypt archive: %w", err)
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

func (s *AppState) defaultOutputPathForEncrypt(inPath string) string {
	return inPath + ".hadescrypt"
}

func (s *AppState) defaultOutputPathForDecrypt(inPath string) string {
	if strings.HasSuffix(strings.ToLower(inPath), ".hadescrypt") {
		return strings.TrimSuffix(inPath, filepath.Ext(inPath))
	}
	return inPath + ".dec"
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