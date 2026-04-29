import os
import shutil
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
    existing_entries = {}
    if os.path.exists(output_file):
        backup_file = "{}.{}.bak".format(
            output_file,
            datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ"),
        )
        shutil.copy2(output_file, backup_file)
        print("Backed up {} to {}".format(output_file, backup_file))

        with open(output_file, "r") as f:
            existing_data = json.load(f)
        for entry in existing_data.get("files", []):
            existing_entries[entry["name"]] = entry

    files_data = {
        "version": 1,
        "files": [],
    }

    seen_names = set()
    for root, _, filenames in os.walk(directory):
        for filename in sorted(filenames):
            if not filename.endswith(".qcow2"):
                continue
            if filename in seen_names:
                continue
            seen_names.add(filename)

            if filename in existing_entries:
                files_data["files"].append(existing_entries[filename])
                continue

            filepath = os.path.join(root, filename)
            print("Hashing new file {}".format(filename))
            file_data = {
                "name": filename,
                "signed": True,
                "hash": md5_hash(filepath),
                "last_modified": last_modified_time(filepath),
            }

            files_data["files"].append(file_data)

    for name, entry in existing_entries.items():
        if name not in seen_names:
            files_data["files"].append(entry)

    files_data["files"].sort(key=lambda e: e["name"])

    with open(output_file, "w") as f:
        json.dump(files_data, f, indent=4)

create_files_json(os.getcwd(), "files.json")
