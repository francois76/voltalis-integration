# Voltalis Integration - Documentation Agent

Ce document sert de référence rapide pour comprendre l'architecture et les concepts de ce projet.

## 🎯 Objectif du Projet

Cette application Go fait le pont entre **Home Assistant** (via MQTT) et l'**API Voltalis** pour le contrôle de radiateurs connectés. L'objectif est de mapper la logique Voltalis vers les entités et concepts Home Assistant.

## 🏗️ Architecture Globale

```
┌─────────────────┐       MQTT        ┌─────────────────────────┐       HTTP        ┌─────────────────┐
│  Home Assistant │◄─────────────────►│   Voltalis Integration  │◄────────────────►│   API Voltalis  │
│     (IHM)       │                   │        (Go App)         │                  │  myvoltalis.com │
└─────────────────┘                   └─────────────────────────┘                  └─────────────────┘
```

## 📁 Structure des Packages

### `/internal/api` - Client API Voltalis
- **`client.go`** : Client HTTP avec authentification Bearer token, méthodes `get()` et `put()`
- **`methods.go`** : Méthodes métier (GetMe, GetAppliances, GetPrograms, EnableQuickSetting, etc.)
- **`structs.go`** : Structures de données correspondant aux réponses API Voltalis

### `/internal/mqtt` - Intégration MQTT Home Assistant
Package principal pour créer et gérer les entités Home Assistant via MQTT Discovery.

- **`client.go`** : Client MQTT avec StateManager pour la gestion d'état
- **`controller.go`** : Entité "Controleur" globale (mode, durée, programme)
- **`heater.go`** : Entités "Climate" pour chaque radiateur
- **`structs.go`** : Payloads de configuration MQTT Discovery (Climate, Select, Sensor, Button)
- **`listen.go`** : Listeners pour les changements d'état depuis HA
- **`publish.go`** : Publication vers MQTT
- **`state_publisher.go`** : Détection de changements d'état (StateManager, diff)
- **`common.go`** : Factories pour générer les payloads de configuration
- **`enums.go`** : Constantes (HeaterMode, HeaterPresetMode, HeaterAction, durées)

### `/internal/transform` - Synchronisation bidirectionnelle
- **`voltalis_to_ha.go`** : Sync Voltalis → Home Assistant (lecture périodique via scheduler)
- **`ha_to_voltalis.go`** : Sync Home Assistant → Voltalis (écoute changements MQTT)
- **`sync_programs.go`** : Synchronisation des programmes disponibles

### `/internal/state` - Gestion d'état
- **`state.go`** : Structures représentant l'état actuel (ResourceState, ControllerState, HeaterState)

### `/internal/scheduler` - Tâches planifiées
- **`scheduler.go`** : Exécution périodique de la sync Voltalis → HA

### `/internal/config` - Configuration
- **`options.go`** : Chargement des options (credentials MQTT/Voltalis)

## 🔄 Flux de Données

### Voltalis → Home Assistant (Lecture)
1. `scheduler` déclenche `transform.SyncVoltalisHeatersToHA()` périodiquement
2. Appels API Voltalis : `GetAppliances()`, `GetPrograms()`
3. Mapping `api.Appliance` → `state.HeaterState` et `state.ControllerState`
4. Publication MQTT sur les topics de commande (`/set`)

### Home Assistant → Voltalis (Écriture)
1. `mqtt.ListenState()` écoute les changements sur les topics `/set`
2. Mise à jour du `StateManager` avec comparaison de l'état précédent
3. Envoi des changements via le channel `StateChange`
4. `transform.Start()` reçoit les changements et doit appeler l'API Voltalis

## 🎛️ Mapping des Concepts

### Modes Voltalis vs PresetModes HA

| Voltalis Mode | HA PresetMode |
|---------------|---------------|
| CONFORT       | Confort       |
| ECO           | Eco           |
| HORS_GEL      | Hors-Gel      |
| TEMPERATURE   | (mode heat)   |

### Types de Programmation Voltalis

| ProgType | Description                        | Mapping HA             |
|----------|------------------------------------|------------------------|
| USER     | Programme hebdomadaire utilisateur | Programme sélectionné  |
| QUICK    | Mode rapide (shortleave, etc.)     | Mode controller        |
| MANUAL   | Réglage manuel température         | Mode heat + temp       |

### QuickSettings Names

| API Name                   | Signification     |
|----------------------------|-------------------|
| `quicksettings.shortleave` | Absence courte    |
| `quicksettings.athome`     | Présence maison   |
| `quicksettings.longleave`  | Absence longue    |

## 🌐 API Voltalis - Endpoints

Base URL: `https://api.myvoltalis.com`

### Authentification
```
POST /auth/login
Body: { "login": "...", "password": "..." }
Response: { "token": "..." }
```

### Lecture
```
GET /api/account/me                              → User info + default site
GET /api/site/{siteId}/managed-appliance         → Liste des radiateurs
GET /api/site/{siteId}/managed-appliance/{id}    → Détail d'un radiateur
GET /api/site/{siteId}/manualsetting             → Réglages manuels
GET /api/site/{siteId}/programming/program       → Liste des programmes
GET /api/site/{siteId}/consumption/realtime      → Consommation temps réel
```

### Écriture
```
PUT /api/site/{siteId}/programming/program/{programId}
Body: { "id": X, "name": "...", "enabled": true/false }

PUT /api/site/{siteId}/quicksettings/{qsId}
Body: { "name": "quicksettings.xxx", "untilFurtherNotice": true, "appliancesSettings": [...], "enabled": true }

PUT /api/site/{siteId}/quicksettings/{qsId}/enable
Body: { "enabled": true/false }

PUT /api/site/{siteId}/manualsetting/{manualSettingId}
Body: { "enabled": true, "idAppliance": X, "untilFurtherNotice": false, "isOn": true, "mode": "ECO", "endDate": "2025-12-08T23:20:34", "temperatureTarget": 20 }
```

## 🏠 Entités Home Assistant Créées

### Par Radiateur (Heater)
- **Climate** : Contrôle température + mode (off/auto/heat) + preset
- **Select "Durée"** : Durée du mode manuel
- **Sensor "Durée mode"** : Affichage de la durée restante

### Controleur Global
- **Select "Mode"** : Eco / Confort / Hors-Gel / Aucun mode
- **Select "Durée"** : Durée d'application du mode
- **Select "Programme"** : Programme hebdomadaire actif
- **Button "Refresh"** : Forcer la resynchronisation

## 📝 Topics MQTT

Pattern: `voltalis/{identifier}/{get|set}`

Exemples:
- `voltalis/voltalis_controller_mode/set` - Commande mode controller
- `voltalis/voltalis_controller_mode/get` - État mode controller
- `voltalis/voltalis_heater_1534507_mode/set` - Commande mode radiateur
- `voltalis/voltalis_heater_1534507_preset_mode/get` - État preset radiateur

## 🔑 Points d'Attention

1. **StateManager** : Utilise un système de hash + diff pour détecter uniquement les vrais changements
2. **Dual Topics** : Chaque entité a un topic `/set` (commande) et `/get` (état)
3. **MQTT Discovery** : Les configs sont publiées sous `homeassistant/{component}/...`
4. **Site ID** : Récupéré automatiquement via `/api/account/me` → `defaultSite.id`

## ⚠️ Pièges de l'API Voltalis (IMPORTANT)

### 1. Le champ `temperatureTarget` est TOUJOURS présent
L'API Voltalis renvoie **toujours** une valeur `temperatureTarget` même quand le mode est ECO/CONFORT/HORS_GEL. **Ne pas se fier à ce champ pour déterminer le mode !** C'est le champ `mode` qui fait foi.

### 2. Le champ `mode` dans ManualSetting détermine le type de contrôle
| Mode API | Signification |
|----------|---------------|
| `CONFORT` | Preset Confort (ignore temperatureTarget) |
| `ECO` | Preset Eco (ignore temperatureTarget) |
| `HORS_GEL` | Preset Hors-Gel (ignore temperatureTarget) |
| `TEMPERATURE` | Température personnalisée (utilise temperatureTarget) |

### 3. Types de programmation (ProgType)
| ProgType | Description | Mode HA correspondant |
|----------|-------------|----------------------|
| `USER` | Programme hebdomadaire actif | `auto` |
| `QUICK` | QuickSetting actif (absence courte, etc.) | Preset selon le quicksetting |
| `MANUAL` | ManualSetting actif (pilotage manuel) | `heat` si TEMPERATURE, sinon preset |
| `DEFAULT` | Aucun programme/setting actif | `auto` |

### 4. Le champ `IsOn` contrôle l'extinction
- `IsOn: true` = radiateur actif (chauffe selon le mode)
- `IsOn: false` = radiateur éteint (mode `off` dans HA)

### 5. Format de date pour `endDate`
Format attendu : `2006-01-02T15:04:05` (sans timezone)

## 🔧 Logique de Synchronisation

### HA → Voltalis (ha_to_voltalis.go)

**Ordre de priorité pour déterminer l'action :**
1. **Mode `off`** → `UpdateManualSetting` avec `IsOn: false`
2. **Mode `auto` sans changement de preset** → Désactiver le ManualSetting (`Enabled: false`)
3. **Changement de preset** (ECO/CONFORT/HORS_GEL) → `UpdateManualSetting` avec le mode correspondant
4. **Mode `heat`** → `UpdateManualSetting` avec `mode: "TEMPERATURE"` et la température

**Important :** Quand on détecte un changement de preset, récupérer la NOUVELLE valeur depuis `changes["PresetMode"]`, pas depuis `heaterState.PresetMode` (qui peut être l'ancienne valeur).

### Voltalis → HA (voltalis_to_ha.go)

**Publication MQTT :**
- Toujours publier le `mode` ET le `preset` (pas l'un OU l'autre)
- Cela permet au StateManager de toujours avoir les deux valeurs à jour

### Gestion du Scheduler

Les handlers dans `ha_to_voltalis.go` retournent `(bool, error)` :
- `true` = des changements ont été appliqués côté Voltalis → déclencher `scheduler.Trigger()` pour resync
- `false` = pas de changement → ne pas déclencher le scheduler

**Ignorer les changements au démarrage :** Vérifier `changes["initial_state"]` pour éviter d'appeler l'API lors de l'initialisation.

## 🚧 TODO / En cours

- [x] Implémentation de `ha_to_voltalis.go` pour appeler les APIs de modification
- [x] Gestion des durées avec calcul de endDate
- [x] Gestion des programmes (activation/désactivation)
- [x] Gestion des quicksettings globaux (mode controller)
- [x] Gestion des manualsettings pour radiateur individuel
- [x] Gestion du mode off (extinction radiateur)
- [x] Gestion du retour au mode auto (désactivation manualSetting)
- [x] Correction du mapping ProgType MANUAL → preset vs température
- [ ] Tests automatisés

## 🧪 Test

```bash
cd test
docker-compose up -d  # Lance Home Assistant + Mosquitto
cd ../voltalis
go run ./cmd/voltalis/main.go
```

## 📋 Exemple de requête ManualSetting fonctionnelle

```bash
# Mettre un radiateur en mode ECO
curl 'https://api.myvoltalis.com/api/site/{siteId}/manualsetting/{manualSettingId}' \
  -X 'PUT' \
  -H 'Authorization: Bearer {token}' \
  -H 'Content-Type: application/json' \
  --data-raw '{"enabled":true,"idAppliance":1534550,"untilFurtherNotice":false,"isOn":true,"mode":"ECO","endDate":"2025-12-08T23:07:40","temperatureTarget":18}'
```

Note : `temperatureTarget` est ignoré quand `mode` est ECO/CONFORT/HORS_GEL, mais doit quand même être présent dans la requête.
