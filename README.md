# 🎮 Power4 Web

Jeu de Puissance 4 en ligne développé en Go.

## 🚀 Installation
```bash
go run main.go
```

Ouvrir `http://localhost:8088`

## 🎯 Comment Jouer

1. Saisir les pseudos (optionnel)
2. Cliquer sur une case pour jouer dans cette colonne
3. Aligner 4 jetons pour gagner !

## 🏗️ Technologies

- **Backend** : Go 1.25
- **Frontend** : HTML + CSS (Go templates)
- **Design** : CSS 

## 📁 Structure
```
power4-web/
├── main.go              # Serveur HTTP
├── game/
│   ├── board.go         # Logique jeu
│   └── win.go           # Détection victoire
├── templates/
│   ├── home.html        # Page accueil
│   └── game.html        # Interface jeu
└── static/
    └── style.css        # Styles
```

## ✨ Fonctionnalités

- ✅ Détection victoire 4 directions
- ✅ Match nul
- ✅ Pseudos personnalisables
- ✅ Scores cumulatifs
- ✅ Animations fluides
- ✅ Responsive design
- ✅ Highlight jetons gagnants

## 👥 Auteurs

Florence Kore-Belle \n
Sarah Bouhadra
Marly Dedjiho
## 📄 Licence

Projet Power4 - B1 Ynov