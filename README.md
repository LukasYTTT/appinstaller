# 📦 AppInstall

[![Release](https://img.shields.io/github/v/release/LukasYTTT/appinstaller?style=flat-square)](https://github.com/LukasYTTT/appinstaller/releases)
[![License](https://img.shields.io/github/license/LukasYTTT/appinstaller?style=flat-square)](LICENSE)

**AppInstall** ist ein moderner, benutzerfreundlicher AppImage-Installer und Manager für Linux. Er ermöglicht es dir, AppImages mit nur wenigen Klicks in dein System zu integrieren (Startmenü-Einträge, Icons, etc.) und sie einfach und sauber wieder zu entfernen.

---

## ✨ Features

- **🎨 Modernes UI**: Ein elegantes Interface mit Dark Mode und Glassmorphism für ein erstklassiges Erlebnis.
- **⚡ Schnelle Integration**: Erstellt automatisch `.desktop`-Dateien und extrahiert passende Icons aus den AppImages.
- **📱 App-Verwaltung**: Übersichtliche Liste aller installierten AppImages mit Ein-Klick-Deinstallation.
- **💻 Terminal-Power**: Volle Funktionalität über die Kommandozeile für Automatisierung und Power-User.
- **⚙️ Hochgradig Anpassbar**: Konfigurierbare Desktop-Pfade, Icon-Einstellungen und Kategorien.

---

## 🚀 Installation

### 📦 Paketmanager (Empfohlen)

Die App wird primär über Paketmanager verteilt, um eine saubere Systemintegration zu gewährleisten.

#### Arch Linux (AUR)
```bash
# Nutze einen AUR-Helper deiner Wahl, z.B. yay:
yay -S appinstall
```

#### Flatpak (Flathub)
```bash
flatpak install flathub io.github.LukasYTTT.appinstall
```

---

### 🛠️ Aus dem Quellcode bauen

Falls du die neueste Version selbst kompilieren möchtest:

1. **Abhängigkeiten**: Stelle sicher, dass `go`, `webkit2gtk`, `gtk3` und `zenity` installiert sind.
2. **Klonen & Bauen**:
   ```bash
   git clone https://github.com/LukasYTTT/appinstaller.git
   cd appinstaller
   make
   ```
3. **Ausführen**:
   ```bash
   ./appinstall
   ```

---

## 🛠️ Benutzung (CLI)

Obwohl die GUI der einfachste Weg ist, bietet `appinstall` eine mächtige CLI:

```bash
# GUI starten (Standard ohne Argumente)
appinstall

# Ein AppImage installieren
appinstall install --appimage ~/Downloads/MyApp.AppImage --name "Mein Tool"

# Installierte Apps auflisten
appinstall uninstall --list

# Eine App sauber deinstallieren
appinstall uninstall --name "Mein Tool"
```

---

## 🔧 Technischer Stack

- **Backend**: [Go](https://go.dev/) mit CGO
- **Frontend**: HTML5, CSS3 (Modern Vanilla), JavaScript
- **Display Engine**: [Webkit2GTK](https://webkitgtk.org/) via `webview_go`
- **Dialoge**: [Zenity](https://help.gnome.org/users/zenity/stable/)

---

## 📜 Lizenz

Dieses Projekt steht unter der **MIT-Lizenz**. Siehe [LICENSE](LICENSE) für Details.

---

Entwickelt mit ❤️ von **LukasYTTT**
