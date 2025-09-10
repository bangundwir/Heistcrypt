💡 Ide Aplikasi Hadscript
Teknologi

Bahasa: Golang

GUI: Fyne

Kriptografi:

AES-256 GCM (default)

XChaCha20-Poly1305 (opsi tambahan)

Argon2id untuk key derivation

HMAC-SHA3 untuk autentikasi

Fitur Utama

Drag & Drop file untuk enkripsi/dekripsi

Enkripsi AES-256 GCM dengan password

Dekripsi cepat dan mudah

Progress bar proses enkripsi/dekripsi

Notifikasi GUI (success/error)

Menampilkan ukuran file sebelum dan sesudah enkripsi/dekripsi

Fitur Lanjutan

Password Generator (custom panjang & karakter)

Password Strength Meter (indikator kuat/lemah)

Keyfiles (single/multiple + urutan)

Mode Paranoid (AES + XChaCha20 + HMAC-SHA3)

Reed-Solomon Error Correction

Force Decrypt

Split into Chunks (MB/GB/TB)

Compress Files (Deflate)

Deniability Mode

Recursive Encrypt/Decrypt

Desain GUI (Fyne)
Tampilan Utama

Header: Logo + Nama aplikasi (Hadscript)

Drag & Drop Area: kotak besar di tengah

File Info Panel: menampilkan nama file, ukuran file, lokasi

Password Input: field password dengan tombol Generate

Password Strength Meter: bar warna menunjukkan kekuatan password

Buttons:

Encrypt (primary)

Decrypt (secondary)

Progress Bar: indikator status proses

Status Bar: pesan hasil (success/error)

Panel Advanced (Expandable/Tab)

Checkbox:

Enable Keyfiles

Enable Paranoid Mode

Enable Reed-Solomon

Force Decrypt

Split into Chunks

Compress Files

Deniability Mode

Recursive Mode

Input tambahan:

Keyfile selector

Chunk size selector

Password generator options (length, charset)

Mockup GUI (Wireframe Text)
+------------------------------------------------+
|                Hadscript (Logo)                |
+------------------------------------------------+
| [ Drag & Drop your files here ]                |
| File: document.pdf   Size: 2.3 MB              |
|                                                |
| Password: [ ************* ]  (Generate)        |
| Strength: [███████-----]  Strong               |
|                                                |
| [ Encrypt ]     [ Decrypt ]                    |
|                                                |
| Progress: [======================     ]  75%   |
| Status: File encrypted successfully            |
+------------------------------------------------+
| Advanced Options ▼                             |
|  [ ] Use Keyfiles   [ Select Keyfile... ]      |
|  [ ] Paranoid Mode                            |
|  [ ] Reed-Solomon ECC                        |
|  [ ] Force Decrypt                           |
|  [ ] Split into Chunks [Size: ___ MB]         |
|  [ ] Compress Files                          |
|  [ ] Deniability Mode                        |
|  [ ] Recursive Mode                          |
+------------------------------------------------+
