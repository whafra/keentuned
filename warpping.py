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
    version_in_spec, _ = warppingCheck()
    if os.path.exists("keentuned-{}".format(version_in_spec)):
        os.system("rm -rf keentuned-{}".format(version_in_spec))
    os.system("mkdir keentuned-{}".format(version_in_spec))
    os.system("go mod vendor")

    os.system("cp -r cli keentuned-{}".format(version_in_spec))
    os.system("cp -r daemon keentuned-{}".format(version_in_spec))
    os.system("cp -r docs keentuned-{}".format(version_in_spec))
    os.system("cp -r vendor keentuned-{}".format(version_in_spec))
    os.system("cp go.mod keentuned-{}".format(version_in_spec))
    os.system("cp go.sum keentuned-{}".format(version_in_spec))
    os.system("cp keentuned.conf keentuned-{}".format(version_in_spec))
    os.system("cp keentuned.service keentuned-{}".format(version_in_spec))
    os.system("cp LICENSE keentuned-{}".format(version_in_spec))
    os.system("cp Makefile keentuned-{}".format(version_in_spec))
    os.system("cp README.md keentuned-{}".format(version_in_spec))

    os.system("tar -cvzf keentuned-{}.tar.gz keentuned-{}".format(version_in_spec, version_in_spec))