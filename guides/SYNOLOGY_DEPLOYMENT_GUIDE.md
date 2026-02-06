# Guía de Despliegue en Synology NAS

Construir en tu PC e Importar
Esta es la más segura para Synology:

# En tu PC, en c:\temp\projects\TheCampaign
docker build -t thecampaign:latest -f dockerfile .
docker save thecampaign:latest -o thecampaign.tar

Luego:

Sube thecampaign.tar al NAS (File Station)
Container Manager → Imagen → Agregar → Agregar desde archivo
Selecciona thecampaign.tar
Ejecuta el contenedor manualmente desde la GUI

Al ejecutar el container importante mapear los puertos 8080 al local 8081 or whatever



Configuraciones para acceder desde fuera

- Reverse proxy en NAS