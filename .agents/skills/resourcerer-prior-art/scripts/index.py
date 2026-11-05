"""Build INDEX.md for each product from its llms.txt, linking the local sanitized files.

usage: python3 -I index.py <a.llms.txt> <b.llms.txt> <out_dir>
"""
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import sanitize as s  # noqa: E402

a_llms, b_llms, out = sys.argv[1:4]


def title_of(path):
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            if line.startswith("# "):
                return line[2:].strip()
    return os.path.basename(path)[:-3]


def build(llms, prefix, subs, link_re, to_file, heading):
    dest = os.path.join(out, prefix)
    local = {f[:-3] for f in os.listdir(dest) if f.endswith(".md") and f != "INDEX.md"}
    seen = set()
    lines = [f"# {heading}", "", "Local copy of the product docs, one file per page. Grep this directory; open a page by file name.", ""]
    item = re.compile(r"^(\s*)- \[(?P<t>[^\]]+)\]\((?P<u>[^)]+)\)(?P<d>.*)$")
    with open(llms, encoding="utf-8") as fh:
        for raw in fh:
            raw = raw.rstrip("\n")
            if raw.startswith("## "):
                lines += ["", s.sub_all(raw.replace("Guides: ", ""), subs), ""]
                continue
            m = item.match(raw)
            if not m:
                continue
            lm = link_re.search("(" + m.group("u").removesuffix(".md").removesuffix("index") + ")")
            if not lm:
                continue
            name = s.sub_all(to_file(lm.group("path")), subs)
            if name not in local:
                continue
            seen.add(name)
            desc = s.sub_all(m.group("d"), subs)
            lines.append(f"{m.group(1)}- [{s.sub_all(m.group('t'), subs)}]({name}.md){desc}")
    rest = sorted(local - seen)
    if rest:
        lines += ["", "## Other pages", ""]
        lines += [f"- [{title_of(os.path.join(dest, n + '.md'))}]({n}.md)" for n in rest]
    with open(os.path.join(dest, "INDEX.md"), "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines) + "\n")
    print(prefix, len(seen), "indexed,", len(rest), "other")


build(a_llms, "atlas", s.A_SUBS, s.A_LINK, lambda p: p, "Atlas docs")
build(b_llms, "orion", s.B_SUBS, s.B_LINK, lambda p: p.strip("/").replace("/", "__") or "home", "Orion docs")
