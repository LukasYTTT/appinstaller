# Anleitung: AppInstall offiziell einreichen (Flathub & AUR)

Diese Anleitung basiert auf den offiziellen Flathub-Richtlinien, um eine reibungslose Annahme deiner App sicherzustellen.

## 1. Lokale Verifizierung (Vor der Einreichung)

Bevor du den Pull Request bei Flathub öffnest, solltest du dein Manifest lokal prüfen und bauen.

### Werkzeuge installieren
Stelle sicher, dass du `flatpak-builder` installiert hast:
```bash
sudo pacman -S flatpak-builder
```
Installiere die Flathub Build-Tools:
```bash
flatpak install -y flathub org.flatpak.Builder
```

### Den Linter ausführen
Der Linter prüft, ob dein Manifest alle Regeln erfüllt:
```bash
flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest assets/io.github.LukasYTTT.appinstall.yml
```

### Lokal bauen und testen
```bash
# Bauen und installieren
flatpak run --command=flathub-build org.flatpak.Builder --install assets/io.github.LukasYTTT.appinstall.yml

# App starten
flatpak run io.github.LukasYTTT.appinstall
```

---

## 2. Einreichung bei Flathub (Offizieller Weg)

1. **Forke das Repo**: Gehe zu [github.com/flathub/flathub](https://github.com/flathub/flathub) und erstelle einen Fork.
2. **Repository lokal vorbereiten**:
   ```bash
   git clone --branch=new-pr git@github.com:DEIN_USERNAME/flathub.git
   cd flathub
   git checkout -b io.github.LukasYTTT.appinstall new-pr
   ```
3. **Manifest hinzufügen**:
   - Erstelle einen Ordner `io.github.LukasYTTT.appinstall`.
   - Kopiere deine Datei `assets/io.github.LukasYTTT.appinstall.yml` dort hinein.
4. **Push & Pull Request**:
   - Pushe deinen Branch zu GitHub.
   - Öffne einen Pull Request gegen den **`new-pr`** Zweig von `flathub/flathub`.
   - Der Titel des PRs sollte sein: `Add io.github.LukasYTTT.appinstall`.

---

## 3. Einreichung beim AUR (Arch Linux)

Folge diesen Schritten, um die App im Arch User Repository zu veröffentlichen:

1. **Repo klonen**: `git clone ssh://aur@aur.archlinux.org/appinstall.git`
2. **Dateien kopieren**: Kopiere das `PKGBUILD` aus deinem Projekt in das geklonte Repo.
3. **Metadaten generieren**: `makepkg --printsrcinfo > .SRCINFO`
4. **Push**: `git add PKGBUILD .SRCINFO && git commit -m "Initial release" && git push`

---

> [!IMPORTANT]
> Achte darauf, dass du auf GitHub ein **Release mit dem Tag `v1.0.0`** erstellt hast, bevor du den Linter ausführst, da die Paketmanager den Quellcode von dort laden möchten.
