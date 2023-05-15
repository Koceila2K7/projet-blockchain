#!/bin/bash

# Créer un conteneur Docker pour la compilation
docker run --rm -v $(pwd):/app -w /app golang:1.18 go build -o blockchain-m2isd

# Vérifier si la compilation a réussi
if [ $? -eq 0 ]; then
    echo "La compilation a réussi. Le binaire a été généré : blockchain-m2isd"
else
    echo "La compilation a échoué."
fi
