import os
import sys
import shutil

vendor_id = "0x1d6b"  # Linux Foundation
product_id = "0x0104"  # Multifunction Composite Gadget
serialnumber = "rpi0w_000004"
manufacturer = "Kosci"
product = "Hane control device"
lang = "0x409"  # english
path = "/sys/kernel/config/usb_gadget/" +product.lower().replace(" ", "_")

functions = [
    {
        "name": "hid.usb0",
        "protocol": "1",
        "subclass": "1",
        "report_length": "8",
        "report_desc": [
            0x05, 0x01,  # USAGE_PAGE (Generic Desktop)
            0x09, 0x06,  # USAGE (Keyboard)
            0xa1, 0x01,  # COLLECTION (Application)
            0x05, 0x07,  # USAGE_PAGE (Keyboard)
            0x19, 0xe0,  # USAGE_MINIMUM (Keyboard LeftControl)
            0x29, 0xe7,  # USAGE_MAXIMUM (Keyboard Right GUI)
            0x15, 0x00,  # LOGICAL_MINIMUM (0)
            0x25, 0x01,  # LOGICAL_MAXIMUM (1)
            0x75, 0x01,  # REPORT_SIZE (1)
            0x95, 0x08,  # REPORT_COUNT (8)
            0x81, 0x02,  # INPUT (Data,Var,Abs)
            0x95, 0x01,  # REPORT_COUNT (1)
            0x75, 0x08,  # REPORT_SIZE (8)
            0x81, 0x01,  # INPUT (Cnst,Var,Abs) // 0x03
            0x95, 0x05,
            0x75, 0x01,
            0x05, 0x08,
            0x19, 0x01,
            0x29, 0x05,
            0x91, 0x02,
            0x95, 0x01,
            0x75, 0x03,
            0x91, 0x01,  # 0x03
            0x95, 0x06,  # REPORT_COUNT (6)
            0x75, 0x08,  # REPORT_SIZE (8)
            0x15, 0x00,  # LOGICAL_MINIMUM (0)
            0x25, 0x65,  # LOGICAL_MAXIMUM (101)
            0x05, 0x07,  # USAGE_PAGE (Keyboard)
            0x19, 0x00,  # USAGE_MINIMUM (Reserved (no event indicated))
            0x29, 0x65,  # USAGE_MAXIMUM (Keyboard Application)
            0x81, 0x00,  # INPUT (Data,Ary,Abs)
            0xc0
        ],
    }
]

def fileputcontent(filename, content, mode="w"):
    with open(filename, mode) as fp:
        if type(content) == str:
            fp.write(content)
        if type(content) == list:
            fp.write(bytearray(content))


def create_dirs(path):
    if not os.path.exists(path):
        os.makedirs(path)

    if not os.path.exists(path + "/strings/" + lang):
        os.makedirs(path + "/strings/" + lang)

    if not os.path.exists(path + "/configs/c.1/strings/" + lang):
        os.makedirs(path + "/configs/c.1/strings/" + lang)


def create_functions(path, functions):
    for function in functions:
        if not os.path.exists(path + "/functions/" + function["name"]):
            os.makedirs(path + "/functions/" + function["name"])
        fileputcontent(path + "/functions/" + function["name"] + "/protocol", function['protocol'])
        fileputcontent(path + "/functions/" + function["name"] + "/subclass", function['subclass'])
        fileputcontent(path + "/functions/" + function["name"] + "/report_length", function['report_length'])
        fileputcontent(path + "/functions/" + function["name"] + "/report_desc", function['report_desc'], "wb")
        if not os.path.exists(path + "/configs/c.1/" + function["name"]):
            os.symlink(path + "/functions/" + function["name"], path + "/configs/c.1/" + function["name"], True)


def enable():
    tmp = os.listdir("/sys/class/udc")
    print(tmp)
    fileputcontent(path + "/UDC", tmp[0])


def disable():
    fileputcontent(path + "/UDC", "")


def install():
    create_dirs(path)
    fileputcontent(path + "/idVendor", vendor_id)
    fileputcontent(path + "/idProduct", product_id)
    fileputcontent(path + "/bcdDevice", "0x0100")
    fileputcontent(path + "/bcdUSB", "0x0200")

    fileputcontent(path + "/strings/" + lang + "/serialnumber", serialnumber)
    fileputcontent(path + "/strings/" + lang + "/manufacturer", manufacturer)
    fileputcontent(path + "/strings/" + lang + "/product", product)

    fileputcontent(path + "/configs/c.1/strings/" + lang + "/configuration", "Config 1: ECM network")
    fileputcontent(path + "/configs/c.1/MaxPower", "250")

    create_functions(path, functions)

    enable()


def remove():
    shutil.rmtree(path)


if len(sys.argv) == 1:
    print("Usage: install, remove, enable, disable")
elif sys.argv[1] == "install":
    install()
elif sys.argv[1] == "remove":
    remove()
elif sys.argv[1] == "disable":
    disable()
elif sys.argv[1] == "enable":
    enable()
else:
    print("No idea")