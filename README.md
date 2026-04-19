# 📦 AppInstall

**AppInstall** ist ein moderner, benutzerfreundlicher AppImage-Installer und Manager für Linux. Er ermöglicht es dir, AppImages mit nur wenigen Klicks in dein System zu integrieren (Startmenü-Einträge, Icons, etc.) und sie einfach wieder zu deinstallieren.

![Premium Design Mockup](https://raw.githubusercontent.com/LukasYTTT/appinstaller/main/assets/preview.png) *(Platzhalter für zukünftigen Screenshot)*

## ✨ Features

- **Intuitive GUI**: Ein modernes Interface (Dark Mode, Glassmorphism) für eine nahtlose Benutzererfahrung.
- **Smart Integration**: Erstellt automatisch `.desktop`-Dateien im Startmenü und lädt/extrahiert Icons.
- **App Management**: Behalte den Überblick über deine installierten AppImages und deinstalliere sie sicher.
- **Integrierte CLI**: Funktioniert auch direkt im Terminal für Power-User.
- **Anpassbar**: Wähle eigene Icons, Kategorien und Desktop-Pfade.

## 🚀 Installation

### Aus dem Quellcode bauen
1. Klone das Repository:
   ```bash
   git clone https://github.com/LukasYTTT/appinstaller.git
   cd appinstaller
   ```
2. Baue die Anwendung:
   ```bash
   make
   ```
3. Installiere sie systemweit:
   ```bash
   sudo make install
   ```

### Arch Linux (AUR)
*Demnächst verfügbar.*

### Flatpak
*Demnächst verfügbar auf Flathub.*

## 🛠️ Benutzung (CLI)

```bash
# GUI starten
appinstall

# AppImage installieren
appinstall install --appimage ~/Downloads/MyApp.AppImage

# App deinstallieren
appinstall uninstall --name "MyApp"
```

## 📜 Lizenz
Dieses Projekt steht unter der **MIT-Lizenz**. Siehe [LICENSE](LICENSE) für Details.

---
Entwickelt mit ❤️ von LukasYTTT
