# Setup for Raspberry PI

## OS Setup

### In Raspbian GUI

- Setup Wifi
- Enable SSH
- Set to boot to CLI
- Reboot

## First Boot

```bash
# change default password
passwd

# install packages
sudo apt-get update
sudo apt-get install tmux vim
```

### Pi Zero / Raspbian Lite

On Mac, enable Internet Connection sharing from Wifi to USB/Ethernet adapter

Enable ssh on Pi Zero:

```bash
sudo raspi-config
```

```bash
sudo apt-get install git
```

### Pi 3 / Raspbian

```bash
# set static IP address
echo $'\n# set static ip\ninterface wlan0\nstatic ip_address=192.168.1.141/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1' | sudo tee --append /etc/dhcpcd.conf
```

## Code

```bash
wget https://storage.googleapis.com/golang/go1.8.3.linux-armv6l.tar.gz -O /tmp/go1.8.3.linux-armv6l.tar.gz
sudo tar -xzf /tmp/go1.8.3.linux-armv6l.tar.gz -C /usr/local

mkdir -p ~/code/go/src/github.com/siggy
git clone https://github.com/siggy/bbox.git ~/code/go/src/github.com/siggy/bbox
```

## Env / bootup


```bash
# git config
git config --global core.editor "vim"
```

```bash
# external sound card
sudo cp ~/code/go/src/github.com/siggy/bbox/rpi/asound.conf /etc/

# set bootup and shell env
cp ~/code/go/src/github.com/siggy/bbox/rpi/.local.bash ~/
echo "[[ -s ${HOME}/.local.bash ]] && source ${HOME}/.local.bash" >> ~/.bashrc
```

# *output of raspi-config after forcing audio to hdmi*
numid=3,iface=MIXER,name='Mic Playback Switch'
  ; type=BOOLEAN,access=rw------,values=1
  : values=on
