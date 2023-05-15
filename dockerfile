# Utilisation de l'image de base Golang
FROM golang:1.18

# Copier le binaire dans le conteneur
COPY blockchain-m2isd /app/

# Définir le répertoire de travail
WORKDIR /app

# Commande à exécuter lorsque le conteneur démarre
CMD ["./blockchain-m2isd"]
