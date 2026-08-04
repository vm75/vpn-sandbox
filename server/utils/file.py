import os

def update_content(content, file_path):
    if os.path.exists(file_path):
        with open(file_path, "r") as f:
            if f.read() == content:
                return False, None
    with open(file_path, "w") as f:
        f.write(content)
    return True, None

def file_exists(filename):
    return os.path.exists(filename)
