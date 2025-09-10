# Comment Detection Feature

HadesCrypt now automatically detects and displays comments from encrypted files when you upload them.

## 🔍 **How Comment Detection Works**

### **File Upload Process**
1. **Drag & Drop File**: Upload any file to HadesCrypt
2. **Automatic Analysis**: App analyzes the file format and content
3. **Comment Extraction**: If it's an encrypted file with comments, extracts them
4. **UI Update**: Comments appear automatically in the comments field

### **Supported File Formats**

#### 🔒 **HadesCrypt Files** (`.hades`, `.hadescrypt`)
- **Comment Storage**: Comments stored in unencrypted header
- **Extraction**: Automatically extracted and displayed
- **Encryption Mode**: Shows which encryption mode was used
- **Full Support**: Complete comment detection and display

#### 🔐 **GnuPG Files** (`.gpg`, `.pgp`)
- **Format Detection**: Automatically detects OpenPGP format
- **Comment Limitation**: GnuPG doesn't store comments in the same way
- **Mode Display**: Shows "GnuPG/OpenPGP" as encryption mode

#### 📁 **Regular Files**
- **No Extraction**: Regular files don't have embedded comments
- **Preserve Existing**: Keeps any comments user has already typed
- **Add New**: User can add comments before encryption

## 🎯 **User Experience**

### **Smart Comment Handling**
```
File Type → Comment Behavior
├── HadesCrypt with comments → Auto-display extracted comments
├── HadesCrypt without comments → Clear comment field  
├── GnuPG files → Keep existing comments (no extraction)
├── Regular files → Keep existing comments
└── Directories → Keep existing comments
```

### **Visual Indicators**
- **🔒 HadesCrypt**: `🔒 Size: 1.2 KB - AES-256-GCM`
- **🔐 GnuPG**: `🔐 Size: 1.5 KB - GnuPG/OpenPGP`
- **📄 Regular**: `Size: 1.0 KB`
- **📁 Directory**: `Folder selected`

## 🛠️ **Technical Implementation**

### **File Analysis Functions**

#### `ExtractCommentsFromFile()`
```go
// Extracts comments from HadesCrypt file header
func ExtractCommentsFromFile(inputPath string) (string, error) {
    // 1. Validate HadesCrypt header ("HAD1")
    // 2. Skip encryption metadata
    // 3. Read comment length (4 bytes)
    // 4. Read comment data
    // 5. Return comment string
}
```

#### `GetFileInfo()`
```go
// Returns comprehensive file information
func GetFileInfo(inputPath string) (map[string]interface{}, error) {
    info := map[string]interface{}{
        "size": fileSize,
        "format": "HadesCrypt|GnuPG/OpenPGP|Unknown",
        "comments": extractedComments,
        "encryption_mode": encryptionMode,
        "encryption_mode_name": humanReadableName,
    }
}
```

#### `IsGnuPGFile()`
```go
// Detects GnuPG/OpenPGP files by extension and content
func IsGnuPGFile(filePath string) bool {
    // 1. Check extensions (.gpg, .pgp)
    // 2. Check OpenPGP packet headers (0x80 bit pattern)
    // 3. Return detection result
}
```

### **UI Integration**

#### `updateFileInfo()`
```go
func (s *AppState) updateFileInfo() {
    // 1. Get file stats
    // 2. Analyze file format
    // 3. Extract comments if available
    // 4. Update UI labels and comment field
    // 5. Preserve user comments for non-encrypted files
}
```

## 📋 **Comment Extraction Process**

### **HadesCrypt File Structure**
```
[4 bytes] Magic Header ("HAD1")
[1 byte]  Version
[1 byte]  Encryption Mode
[16 bytes] Salt
[8 bytes]  Nonce Prefix
[4 bytes]  Chunk Size
[8 bytes]  Total Size
[4 bytes]  Comment Length  ← Comment detection starts here
[N bytes]  Comment Data    ← Extracted and displayed
[...] Encrypted Content
```

### **Extraction Steps**
1. **Header Validation**: Verify "HAD1" magic header
2. **Version Check**: Ensure compatible version
3. **Metadata Skip**: Skip encryption-specific fields
4. **Comment Length**: Read 4-byte comment length
5. **Comment Data**: Read N bytes of comment
6. **Validation**: Check for reasonable comment size (max 1MB)
7. **Display**: Show comment in UI

## 🎨 **User Interface Features**

### **Smart Comment Field Behavior**
- **Auto-populate**: Comments from encrypted files appear automatically
- **Preserve User Input**: Don't overwrite user-typed comments for regular files
- **Clear When Appropriate**: Clear comments for encrypted files without comments
- **Visual Feedback**: Different icons for different file types

### **File Information Display**
```
Before: Size: 1.2 KB
After:  🔒 Size: 1.2 KB - AES-256-GCM
        🔐 Size: 1.5 KB - GnuPG/OpenPGP
```

### **Comment Field Integration**
- **Automatic Population**: Comments appear without user action
- **Editable**: User can still modify extracted comments
- **Preserved**: Comments maintained during encryption/decryption cycle

## 🧪 **Usage Examples**

### **Example 1: Encrypted File with Comments**
1. **Create file**: `important_document.txt`
2. **Add comment**: "Contains quarterly financial report"
3. **Encrypt**: Choose AES-256, encrypt → `important_document.txt.hadescrypt`
4. **Later**: Drag `important_document.txt.hadescrypt` to HadesCrypt
5. **Result**: Comment automatically appears: "Contains quarterly financial report"
6. **UI shows**: `🔒 Size: 2.1 KB - AES-256-GCM`

### **Example 2: GnuPG File**
1. **Encrypt with GnuPG**: `secret.txt` → `secret.txt.gpg`
2. **Upload to HadesCrypt**: Drag `secret.txt.gpg`
3. **Result**: Detected as GnuPG file
4. **UI shows**: `🔐 Size: 1.8 KB - GnuPG/OpenPGP`
5. **Comments**: No automatic extraction (GnuPG limitation)

### **Example 3: Regular File**
1. **Upload**: `document.txt`
2. **Add comment**: "Draft version for review"
3. **Switch files**: Upload different file, then back to `document.txt`
4. **Result**: Comment "Draft version for review" preserved
5. **UI shows**: `Size: 1.0 KB`

## 🔧 **Advanced Features**

### **Encryption Mode Detection**
```go
// Detected modes displayed in UI:
ModeAES256GCM → "AES-256-GCM"
ModeChaCha20 → "ChaCha20-Poly1305"  
ModeParanoid → "Paranoid (AES-256 + ChaCha20)"
ModePostQuantumKyber768 → "Post-Quantum: Kyber-768"
ModeGnuPG → "GnuPG/OpenPGP"
```

### **File Size Formatting**
```go
FormatFileSize(1024) → "1.00 KB"
FormatFileSize(1048576) → "1.00 MB"
FormatFileSize(1073741824) → "1.00 GB"
```

### **Error Handling**
- **Invalid Files**: Graceful fallback to basic file info
- **Corrupted Headers**: Safe error handling without crashes
- **Large Comments**: Sanity check (max 1MB) prevents abuse
- **Permission Issues**: Proper error messages

## 🎯 **Benefits**

### **For Users**
- **Convenience**: No need to remember what encrypted files contain
- **Organization**: Comments help identify file contents
- **Workflow**: Seamless integration with existing workflow
- **Visual Feedback**: Clear indicators for different file types

### **For Security**
- **Metadata Preservation**: Comments stored securely in file header
- **No Data Loss**: Comments preserved through encryption/decryption cycle
- **Format Detection**: Prevents wrong decryption attempts
- **Safe Extraction**: No risk of exposing encrypted content

## 🚀 **Future Enhancements**

### **Planned Features**
- **Comment Editing**: Edit comments in encrypted files without full decryption
- **Metadata Display**: Show creation date, encryption timestamp
- **Batch Analysis**: Analyze multiple files at once
- **Comment Search**: Search through comments in multiple files

### **GnuPG Improvements**
- **ASCII Armor Comments**: Extract comments from ASCII armored files
- **Signature Info**: Display signature information if present
- **Key Information**: Show which keys were used for encryption

---

*HadesCrypt Comment Detection - Making encrypted file management effortless* 💬🔒
