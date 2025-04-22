import os
import subprocess
import json
from datetime import datetime, timezone

def md5_hash(filepath):
    result = subprocess.run(["md5sum", filepath], stdout=subprocess.PIPE)
    return result.stdout.split()[0].decode("utf-8")

def last_modified_time(filepath):
    timestamp = os.path.getmtime(filepath)
    return datetime.fromtimestamp(timestamp,
        tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%S") + "Z"

def create_files_json(directory, output_file):
    files_data = {
        "version": 1,
        "files": [],
    }

    for root, _, filenames in os.walk(directory):
        for filename in sorted(filenames):
            if not filename.endswith(".qcow2"):
                continue

            filepath = os.path.join(root, filename)
            file_data = {
                "name": filename,
                "signed": True,
                "hash": md5_hash(filepath),
                "last_modified": last_modified_time(filepath),
            }

            files_data["files"].append(file_data)

    with open(output_file, "w") as f:
        json.dump(files_data, f, indent=4)

create_files_json(os.getcwd(), "files.json")
