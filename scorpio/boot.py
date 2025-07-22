import usb_cdc
# enable the USB REPL **and** the data channel before code.py runs
usb_cdc.enable(console=True, data=True)
