#!/bin/bash

# Получаем текущую директорию
current_dir=$(pwd)

for file in *.go; do
    [ -f "$file" ] || continue

    filename=$(basename -- "$file")
    basename="${filename%.*}"
    mockdir="mocks_${basename}"

    # Создаём директорию для моков, если её нет
    mkdir -p "$mockdir"

    # Генерация мока
    mockgen -source="$file" -destination="$mockdir/mock_${basename}.go" -package="$mockdir"
done
