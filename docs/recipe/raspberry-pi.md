# Proof of Concept

Raspberry Pi 4b with:

* Reticulum RNode LoRA interface
* Airgradient sensors
  * Polling serial for CO2, NOx, AQI, and temp sensors, or alternatively scraping prometheus endpoint over wifi
* Meshtastic interface
  * [serial modes](https://meshtastic.org/docs/configuration/module/serial/)
* OpenWRT 802.11s mesh router
  * Bridge to other IP networks
* Solar powered

## Setup

### Install OS on Raspberry Pi

This is an exercise left to the reader. You may find [rpi-image-gen](https://github.com/raspberrypi/rpi-image-gen) and the exemplars in in platform/rpi-image-gen helpful.

### Configure Services

Additional steps required:

1. Setup SSH
2. Install container runtime
3. Install kubelet
4. Configure [static pod](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod/) to run multiband

### Power

* Car battery
* Cheap solar controller
* Foldable solar panel
* 12v USB adapter
