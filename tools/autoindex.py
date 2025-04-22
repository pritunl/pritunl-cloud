import os
import sys
from datetime import datetime, timezone

ROOT_DIR = "/mnt/images"
PREFIX = ""
if len(sys.argv) > 1:
    PREFIX = sys.argv[1] + "/"
HTML_HEADER = """<html><head><title>Index of /{relative_path}</title></head>
<body bgcolor="white">
<h1>Index of /{relative_path}</h1><hr><pre><a href="../">../</a>
"""
HTML_FOOTER = """</pre><hr>
</body></html>
"""

def limit_filename_length(filename, max_length=50):
    if len(filename) > max_length:
        return filename[:max_length - 3] + '..>'
    return filename

def generate_directory_listing_html(current_dir, relative_path, dirs, files):
    html = HTML_HEADER.format(relative_path=PREFIX + relative_path)

    for directory in sorted(dirs):
        dir_path = os.path.join(current_dir, directory)
        stat_info = os.stat(dir_path)
        modified_time = datetime.fromtimestamp(stat_info.st_mtime,
            tz=timezone.utc).strftime("%d-%b-%Y %H:%M")
        label = limit_filename_length(directory + '/')
        html += f"<a href=\"{directory}/\">{label}</a>" + \
            f"{' ' * (51 - len(label))}{modified_time}{' ' * 20}-\n"

    for file_name in sorted(files):
        if file_name == "index.html":
            continue
        file_path = os.path.join(current_dir, file_name)
        stat_info = os.stat(file_path)
        modified_time = datetime.fromtimestamp(stat_info.st_mtime,
            tz=timezone.utc).strftime("%d-%b-%Y %H:%M")
        file_size = f"{stat_info.st_size:21}"
        label = limit_filename_length(file_name)
        html += f"<a href=\"{file_name}\">{label}</a>" + \
            f"{' ' * (51 - len(label))}{modified_time}{file_size}\n"

    html += HTML_FOOTER

    return html

def generate_index_files(root_dir):
    for current_dir, subdirs, files in os.walk(root_dir):
        relative_path = os.path.relpath(current_dir, root_dir)
        if relative_path == ".":
            relative_path = ""
        else:
            relative_path += "/"

        html_content = generate_directory_listing_html(current_dir,
            relative_path, sorted(subdirs), sorted(files))
        index_path = os.path.join(current_dir, "index.html")

        with open(index_path, "w") as f:
            f.write(html_content)

        print(f"Generated index.html for: {current_dir}")

generate_index_files(ROOT_DIR)
