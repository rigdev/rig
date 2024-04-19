import sys


def truncate(path: str):
    try:
        with open(path) as f:
            contents = f.read()
            idx = contents.index("## Config")
    except:
        return

    contents = contents[:idx]
    with open(path, "w") as f:
        f.write(contents)


if __name__ == "__main__":
    path = sys.argv[1]
    truncate(path)
