import os
import re  

def warppingCheck():
    with open("keentuned.spec",'r') as f:
        spec = f.read()
        version_in_spec = re.search("Version:        ([\d.]+)\n",spec).group(1)
        release_in_spec = re.search("define anolis_release (\d)\n",spec).group(1)
        print("Get version: {}-{}".format(version_in_spec, release_in_spec))
        
        if re.search(" - {}-{}".format(version_in_spec, release_in_spec), spec):
            print("[OK] check version in changelog at keentuned.spec")
        else:
            print("[Failed] wrong version number in changelog at keentuned.spec")
            return
        
    with open("Makefile",'r') as f:
        makefile = f.read()

        if re.search(version_in_spec, makefile):
            print("[OK] check version in changelog at Makefile")
        else:
            print("[Failed] wrong version number in changelog at Makefile")
            return

    print("Start wrap up of keentune-{}-{}".format(version_in_spec, release_in_spec))
    return version_in_spec, release_in_spec


if __name__ == "__main__":
    version_in_spec, release_in_spec = warppingCheck()
    os.system("go mod vendor")
    os.system("tar -cvzf keentuned-{}.tar.gz cli daemon docs vendor go.mod go.sum keentuned.conf keentuned.service LICENSE Makefile README.md".format(version_in_spec))