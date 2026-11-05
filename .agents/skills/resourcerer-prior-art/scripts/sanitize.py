"""Sanitize downloaded product docs: replace brand names with codenames and rewrite internal links to local files.

usage: python3 -I sanitize.py <raw_a> <raw_b> <out_dir>
"""
import os
import re
import sys

A_SUBS = [
    (r"\b(?:Cast|CAST)\.?\s?AI\b", "Atlas"),
    (r"\bCastAI\b", "Atlas"),
    (r"CASTAI", "ATLAS"),
    (r"Castai", "Atlas"),
    (r"castai", "atlas"),
    (r"\bcast\.ai\b", "atlas.example"),
    (r"\bcast\\\.ai\b", r"atlas\\.example"),
    (r"\bcast-ai\b", "atlas"),
    (r"CASTWARE", "ATLASWARE"),
    (r"\bcast(?=pool|aksrole|workers|[A-Z_])", "atlas"),
    (r"\bCAST(?=_|\b)", "ATLAS"),
    (r"_CAST_", "_ATLAS_"),
    (r"\bcast\b", "atlas"),
    (r"\bCast(?=[A-Z])", "Atlas"),
    (r"\bcast(?=ctl|ware|pilot)", "atlas"),
    (r"\bCast(?=ware|pilot)", "Atlas"),
    (r"\bCast\b", "Atlas"),
    (r"Kvisor", "Sentinel"),
    (r"kvisor", "sentinel"),
    (r"KVISOR", "SENTINEL"),
    (r"Kentroller", "Keeper"),
    (r"kentroller", "keeper"),
    (r"KENTROLLER", "KEEPER"),
]
B_SUBS = [
    (r"\bstormforge\.io\b", "orion.example"),
    (r"Storm\s?Forge", "Orion"),
    (r"Stormforge", "Orion"),
    (r"STORMFORGE", "ORION"),
    (r"stormforge", "orion"),
    (r"CloudBolt", "OrionCorp"),
    (r"Cloudbolt", "OrionCorp"),
    (r"cloudbolt", "orioncorp"),
    (r"Optimize (Live|Pro)", r"Orion \1"),
    (r"\bsf_", "orion_"),
    (r"\*\*sf\*\*", "**orion**"),
]


def sub_all(text, subs):
    for pat, rep in subs:
        text = re.sub(pat, rep, text)
    return text


def slug_a(name):
    return sub_all(name, A_SUBS)


def slug_b(name):
    return sub_all(name, B_SUBS)


def strip_boilerplate(text):
    text = re.sub(r"agentTools:\n(?:  .*\n)+", "", text)
    text = re.sub(r"\n---\n\nSite \[llms\.txt\]\(/llms\.txt\)\n\n---\n", "\n", text)
    return text


def process(out, raw, prefix, subs, slugger, link_re, to_file):
    files = {f[:-3]: slugger(f[:-3]) for f in os.listdir(raw) if f.endswith(".md")}
    local = set(files.values())
    dest = os.path.join(out, prefix)
    os.makedirs(dest, exist_ok=True)

    def relink(m):
        target, anchor = to_file(m.group("path")), m.group("anchor") or ""
        target = slugger(target)
        if target in local:
            return f"({target}.md{anchor})"
        return m.group(0)

    for src, name in files.items():
        with open(os.path.join(raw, src + ".md"), encoding="utf-8") as fh:
            text = fh.read()
        text = strip_boilerplate(text)
        text = link_re.sub(relink, text)
        text = sub_all(text, subs)
        with open(os.path.join(dest, name + ".md"), "w", encoding="utf-8") as fh:
            fh.write(text)
    return len(files)


A_LINK = re.compile(
    r"\((?:https://docs\.cast\.ai)?/(?:castai/)?docs/(?P<path>[A-Za-z0-9._-]+)/?(?P<anchor>#[^)\s]*)?\)"
)
B_LINK = re.compile(
    r"\((?:https://docs\.stormforge\.io)?/docs/(?P<path>[A-Za-z0-9._/-]*?)/?(?:index\.md)?(?P<anchor>#[^)\s]*)?\)"
)

if __name__ == "__main__":
    raw_a, raw_b, out = sys.argv[1:4]
    n_a = process(out, raw_a, "atlas", A_SUBS, slug_a, A_LINK, lambda p: p)
    n_b = process(out, raw_b, "orion", B_SUBS, slug_b, B_LINK, lambda p: p.strip("/").replace("/", "__") or "home")
    print(f"atlas={n_a} orion={n_b}")
