# Portal64 API Demo

Ein vollständiges HTML-Demo-Interface für die Portal64 DWZ Chess Rating System API.

## Übersicht

Diese Demo bietet eine benutzerfreundliche Weboberfläche zum Testen aller API-Endpunkte des Portal64 Systems. Das Design verwendet ein modernes schwarz-weißes Farbschema mit responsivem Layout.

## Dateien-Struktur

```
demo/
├── index.html              # Hauptdashboard und Übersichtsseite
├── players.html            # Player-API Testinterface
├── clubs.html              # Club-API Testinterface  
├── tournaments.html        # Tournament-API Testinterface
├── api-docs.html          # Vollständige API-Dokumentation
├── css/
│   ├── main.css           # Grundlegende Styles und Layout
│   ├── forms.css          # Formular-spezifische Styles
│   └── components.css     # Wiederverwendbare Komponenten
└── js/
    ├── api.js             # API-Client Klasse
    ├── utils.js           # Hilfsfunktionen
    ├── main.js            # Hauptanwendungslogik
    ├── players.js         # Player-spezifische Funktionalität
    ├── clubs.js           # Club-spezifische Funktionalität
    └── tournaments.js     # Tournament-spezifische Funktionalität
```

## Setup und Verwendung

### Voraussetzungen

1. **Portal64 API Server**: Stellen Sie sicher, dass der API-Server läuft:
   ```bash
   cd C:\Users\tranm\work\svw.info\portal64api
   .\server.exe
   ```
   
2. **MySQL Datenbank**: MySQL sollte auf localhost:3306 laufen mit:
   - Username: `root`
   - Password: `` (leer)
   - MySQL-Pfad: `c:\xampp\mysql\bin`

### Demo starten

1. **Öffnen Sie die Demo**: 
   - Navigieren Sie zu `C:\Users\tranm\work\svw.info\portal64api\demo\`
   - Öffnen Sie `index.html` in einem modernen Webbrowser

2. **API-Verbindung prüfen**:
   - Die Startseite führt automatisch einen Health-Check durch
   - Grüner Status = API läuft korrekt
   - Roter Status = Überprüfen Sie die API-Server-Verbindung

## Funktionen

### 🏠 Dashboard (index.html)
- **Health Check**: Automatische API-Statusprüfung
- **Navigation**: Schnellzugriff auf alle Demo-Bereiche  
- **Quick Tests**: Sofortiges Testen häufiger Endpunkte
- **API-Informationen**: Grundlegende Systemdetails

### 👤 Players Interface (players.html)
- **Player-Suche**: Volltext-Suche mit Pagination und Sortierung
- **Einzelspieler-Lookup**: Detaillierte Spielerinformationen per Player-ID
- **Rating-Historie**: Komplette DWZ-Bewertungsverläufe
- **Vereinsspieler**: Alle Spieler eines bestimmten Vereins
- **Format-Optionen**: JSON oder CSV Export

### 🏛️ Clubs Interface (clubs.html)
- **Vereinssuche**: Suche nach Name, Region oder VKZ
- **Alle Vereine**: Vollständige Vereinsliste
- **Vereinsdetails**: Detaillierte Vereinsinformationen per VKZ
- **Filteroptionen**: Region und Bezirk Filter
- **Sortierung**: Nach verschiedenen Kriterien sortierbar

### 🏆 Tournaments Interface (tournaments.html)
- **Turniersuche**: Allgemeine Turniersuche mit Filtering
- **Kommende Turniere**: Turniere mit offener Anmeldung
- **Aktuelle Turniere**: Kürzlich beendete Turniere
- **Datumsbereich**: Turniere in einem bestimmten Zeitraum
- **Turnierdetails**: Vollständige Turnierinformationen

### 📚 API Documentation (api-docs.html)
- **Vollständige Referenz**: Alle Endpunkte dokumentiert
- **Request/Response Beispiele**: Praktische Code-Beispiele
- **Parameter-Details**: Vollständige Parameterbeschreibungen
- **Swagger Integration**: Link zur interaktiven API-Dokumentation
- **Testfunktionen**: Direkte Endpunkt-Tests aus der Dokumentation

## Design-Features

### 🎨 Visuelles Design
- **Modernes schwarz-weiß Theme**: Professionelles Erscheinungsbild
- **Responsive Layout**: Funktioniert auf Desktop, Tablet und Mobile
- **Glassmorphismus**: Moderne UI-Effekte und Übergänge
- **Zugänglichkeit**: Hoher Kontrast und semantisches HTML

### 🔧 Technische Features
- **Tab-basierte Navigation**: Organisierte Benutzeroberfläche
- **Live API-Testing**: Echte API-Aufrufe mit formatierter Ausgabe
- **Error Handling**: Benutzerfreundliche Fehlermeldungen
- **Copy-to-Clipboard**: Einfaches Kopieren von API-Responses
- **Modal-Dialoge**: Detailansichten ohne Seitenwechsel
- **Form-Validierung**: Client-seitige Eingabevalidierung

### ⚡ Performance
- **Modulare JavaScript-Architektur**: Wartbarer und erweiterbarer Code
- **CSS-Optimierung**: Effiziente Stylesheet-Organisation  
- **Lazy Loading**: Tabs werden nur bei Bedarf initialisiert
- **Debounced Search**: Optimierte Suchanfragen

## API-Endpunkte im Detail

### Health Check
- `GET /health` - API-Statusprüfung

### Players
- `GET /api/v1/players` - Spielersuche mit Pagination
- `GET /api/v1/players/{id}` - Einzelspieler-Details
- `GET /api/v1/players/{id}/rating-history` - DWZ-Verlauf
- `GET /api/v1/clubs/{id}/players` - Vereinsspieler

### Clubs  
- `GET /api/v1/clubs` - Vereinssuche mit Filtering
- `GET /api/v1/clubs/all` - Alle Vereine
- `GET /api/v1/clubs/{id}` - Vereinsdetails

### Tournaments
- `GET /api/v1/tournaments` - Turniersuche
- `GET /api/v1/tournaments/recent` - Aktuelle Turniere  
- `GET /api/v1/tournaments/date-range` - Turniere nach Datum
- `GET /api/v1/tournaments/{id}` - Turnierdetails

## Format-Unterstützung

Alle Endpunkte unterstützen mehrere Ausgabeformate:

### JSON (Standard)
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

### CSV Export
- Per `?format=csv` Parameter
- Oder `Accept: text/csv` Header  
- Direkter Download in Browser

## ID-Formate

### Player IDs
- Format: `C####-####`
- Beispiel: `C0101-1014`
- Validierung: Automatische Format-Prüfung

### Club IDs (VKZ)
- Format: `C####`
- Beispiel: `C0101`  
- VKZ = Vereins-Kennziffer

### Tournament IDs
- Format: `C###-###-###`
- Beispiel: `C529-K00-HT1`
- Eindeutige Turnier-Kennungen

## Fehlerbehebung

### API-Server nicht erreichbar
1. Prüfen Sie ob der Server läuft: `http://localhost:8080/health`
2. Starten Sie den Server neu: `.\server.exe`
3. Überprüfen Sie die Port-Konfiguration (Standard: 8080)

### MySQL-Verbindungsfehler
1. Starten Sie XAMPP MySQL
2. Prüfen Sie die Datenbankverbindung
3. Überprüfen Sie Username/Password Konfiguration

### Browser-Kompatibilität
- Moderne Browser erforderlich (Chrome 70+, Firefox 65+, Safari 12+)
- JavaScript muss aktiviert sein
- Cookies für localhost erlauben

## Erweiterung

### Neue Endpunkte hinzufügen
1. Erweitern Sie `js/api.js` um neue API-Methoden
2. Erstellen Sie entsprechende UI-Elemente
3. Implementieren Sie Handler-Funktionen
4. Aktualisieren Sie die Dokumentation

### Styling anpassen  
1. Bearbeiten Sie CSS-Variablen in `css/main.css`
2. Passe Farbschema in `:root` Sektor an
3. Responsive Breakpoints in Media Queries

### Neue Features
1. Folgen Sie der bestehenden Modulstruktur
2. Verwenden Sie etablierte Utility-Funktionen
3. Behalten Sie Konsistenz im Design bei

## Support

Bei Fragen oder Problemen:
1. Überprüfen Sie die Browser-Konsole auf JavaScript-Fehler
2. Testen Sie API-Endpunkte direkt mit cURL oder Postman  
3. Überprüfen Sie die Server-Logs für Backend-Probleme

---

**Hinweis**: Diese Demo ist für Entwicklungs- und Testzwecke gedacht. Für Produktionsumgebungen sollten entsprechende Sicherheitsmaßnahmen implementiert werden.