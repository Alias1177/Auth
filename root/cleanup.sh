#!/bin/bash
# Файл: /root/cleanup.sh

# Записываем текущее использование диска
DISK_USAGE=$(df -h | grep '/dev/sda1' | awk '{print $5}' | sed 's/%//')

# Если использование диска превышает 80%, выполняем очистку
if [ "$DISK_USAGE" -gt 80 ]; then
  echo "Диск заполнен на $DISK_USAGE%. Запускаем очистку..."

  # Очистка Docker
  docker system prune -af --volumes

  # Очистка кэша apt и журналов
  apt-get clean
  apt-get autoremove -y
  journalctl --vacuum-time=3d

  # Удаление старых логов
  find /var/log -type f -name "*.gz" -delete
  find /var/log -type f -name "*.log.*" -delete


  echo "Очистка завершена"
else
  echo "Диск заполнен на $DISK_USAGE%. Очистка не требуется."
fi