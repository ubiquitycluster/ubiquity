import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / 'graphify-out'
MANIFEST = json.loads((OUT / '.graphify_chunks_manifest.json').read_text(encoding='utf-8'))

STOP = {
    'the','and','for','with','from','this','that','into','using','use','are','was','were','has','have','will','shall','must','may','can',
    'kubernetes','cluster','clusters','service','services','system','config','configuration','default','values','template','templates'
}
KEYWORDS = [
    'argocd','helm','kustomize','k3s','cert-manager','vault','external-secrets','longhorn','velero','cilium','metallb','ingress',
    'monitoring','prometheus','grafana','loki','alloy','harbor','gitea','renovate','dependabot','nvidia','gpu','nico','metal3',
    'pxe','ironic','bare-metal','bare metal','terraform','ansible','cloud-init','aws','azure','gcp','openstack','ovh','sandbox',
    'production','backup','restore','security','networkpolicy','policy','storage','postgres','redis','dragonfly','ollama','hajimari'
]
RELATIONS = {
    'argocd': 'references', 'helm': 'implements', 'kustomize': 'implements', 'k3s': 'implements', 'cert-manager': 'references',
    'vault': 'references', 'external-secrets': 'references', 'longhorn': 'references', 'velero': 'references', 'cilium': 'references',
    'metallb': 'references', 'ingress': 'references', 'monitoring': 'conceptually_related_to', 'prometheus': 'references',
    'grafana': 'references', 'loki': 'references', 'alloy': 'references', 'harbor': 'references', 'gitea': 'references',
    'renovate': 'references', 'dependabot': 'references', 'nvidia': 'references', 'gpu': 'conceptually_related_to', 'nico': 'references',
    'metal3': 'references', 'pxe': 'conceptually_related_to', 'ironic': 'references', 'bare-metal': 'conceptually_related_to',
    'bare metal': 'conceptually_related_to', 'terraform': 'implements', 'ansible': 'implements', 'cloud-init': 'implements',
    'aws': 'references', 'azure': 'references', 'gcp': 'references', 'openstack': 'references', 'ovh': 'references',
    'sandbox': 'conceptually_related_to', 'production': 'conceptually_related_to', 'backup': 'conceptually_related_to',
    'restore': 'conceptually_related_to', 'security': 'conceptually_related_to', 'networkpolicy': 'implements',
    'policy': 'conceptually_related_to', 'storage': 'conceptually_related_to', 'postgres': 'references', 'redis': 'references',
    'dragonfly': 'references', 'ollama': 'references', 'hajimari': 'references'
}


def relpath(p: Path) -> str:
    try:
        return str(p.relative_to(ROOT))
    except ValueError:
        return str(p)


def stem_prefix(path: str) -> str:
    p = Path(path)
    parent = p.parent.name
    stem = p.stem
    base = f'{parent}_{stem}' if parent and str(p.parent) != '.' else stem
    return re.sub(r'[^a-z0-9]+', '_', base.lower()).strip('_') or 'root'


def node_id(prefix: str, entity: str) -> str:
    ent = re.sub(r'[^a-z0-9]+', '_', entity.lower()).strip('_') or 'document'
    return f'{prefix}_{ent}'[:160].strip('_')


def label_from_path(path: str) -> str:
    p = Path(path)
    return re.sub(r'[-_]+', ' ', p.stem).strip().title() or p.name


def read_text(path: Path) -> str:
    try:
        return path.read_text(encoding='utf-8', errors='ignore')[:20000]
    except Exception:
        return ''


def file_type(path: str) -> str:
    ext = Path(path).suffix.lower()
    if ext in {'.png','.jpg','.jpeg','.gif','.webp','.svg'}:
        return 'image'
    if ext in {'.go','.py','.sh','.bash','.tf','.json'}:
        return 'code'
    return 'document'


def extract_entities(text: str, path: str):
    found=[]
    low=text.lower()
    path_low=path.lower()
    for kw in KEYWORDS:
        if kw in low or kw in path_low:
            found.append(kw)
    # YAML-ish Kubernetes kinds and obvious headings
    for m in re.finditer(r'(?mi)^\s*(?:kind|name):\s*([A-Za-z0-9_.-]{3,64})\s*$', text):
        val=m.group(1).strip().lower()
        if val not in STOP:
            found.append(val)
    for m in re.finditer(r'(?m)^#{1,3}\s+(.{3,80})$', text):
        val=re.sub(r'[^A-Za-z0-9 -]+','',m.group(1)).strip().lower()
        words=[w for w in val.split() if w not in STOP]
        if words:
            found.append(' '.join(words[:4]))
    # de-dupe preserving order
    out=[]; seen=set()
    for e in found:
        e=e.strip().lower()
        if e and e not in seen:
            seen.add(e); out.append(e)
    return out[:8]


def make_chunk(idx: int):
    filelist = (OUT / f'.graphify_chunk_{idx:02d}.files').read_text(encoding='utf-8').splitlines()
    nodes=[]; edges=[]; hyper=[]; node_ids=set(); concept_nodes={}
    files_by_concept={}
    for raw in filelist:
        path=Path(raw)
        if not path.is_absolute(): path=ROOT/path
        rp=relpath(path)
        prefix=stem_prefix(rp)
        ftype=file_type(rp)
        doc_id=node_id(prefix, 'document')
        nodes.append({
            'id': doc_id,
            'label': label_from_path(rp),
            'file_type': ftype,
            'source_file': rp,
            'source_location': None,
            'source_url': None,
            'captured_at': None,
            'author': None,
            'contributor': None,
        })
        node_ids.add(doc_id)
        text=read_text(path)
        ents=extract_entities(text, rp)
        for ent in ents:
            cid=node_id(prefix, ent)
            if cid not in node_ids:
                nodes.append({
                    'id': cid,
                    'label': ent.replace('-', ' ').title(),
                    'file_type': 'concept',
                    'source_file': rp,
                    'source_location': None,
                    'source_url': None,
                    'captured_at': None,
                    'author': None,
                    'contributor': None,
                })
                node_ids.add(cid)
            edges.append({
                'source': doc_id,
                'target': cid,
                'relation': RELATIONS.get(ent, 'conceptually_related_to'),
                'confidence': 'INFERRED',
                'confidence_score': 0.75 if ent in KEYWORDS else 0.65,
                'source_file': rp,
                'source_location': None,
                'weight': 1.0,
            })
            files_by_concept.setdefault(ent, []).append(cid)
        # explicit path hierarchy concept
        parts=[x for x in Path(rp).parts[:-1] if x not in {'.'}]
        if parts:
            ent=' '.join(parts[:2])
            hid=node_id(prefix, ent)
            if hid not in node_ids:
                nodes.append({'id':hid,'label':ent.title(),'file_type':'concept','source_file':rp,'source_location':None,'source_url':None,'captured_at':None,'author':None,'contributor':None})
                node_ids.add(hid)
            edges.append({'source':doc_id,'target':hid,'relation':'conceptually_related_to','confidence':'INFERRED','confidence_score':0.65,'source_file':rp,'source_location':None,'weight':1.0})
    # cross-file concept hyperedges, capped at 3
    hcount=0
    for ent, ids in files_by_concept.items():
        uniq=[]
        for x in ids:
            if x not in uniq: uniq.append(x)
        if len(uniq)>=3 and hcount<3:
            hyper.append({
                'id': re.sub(r'[^a-z0-9]+','_',f'chunk_{idx}_{ent}_shared_concept'.lower()).strip('_'),
                'label': f'Shared {ent.replace("-", " ").title()} Concept',
                'nodes': uniq[:8],
                'relation': 'participate_in',
                'confidence': 'INFERRED',
                'confidence_score': 0.75,
                'source_file': nodes[0]['source_file'] if nodes else '',
            })
            hcount+=1
    # de-dupe edges
    seen=set(); dedup=[]
    for e in edges:
        key=(e['source'],e['target'],e['relation'],e['source_file'])
        if key not in seen and e['source'] in node_ids and e['target'] in node_ids:
            seen.add(key); dedup.append(e)
    return {'nodes': nodes, 'edges': dedup, 'hyperedges': hyper, 'input_tokens': 0, 'output_tokens': 0}


def main():
    made=[]
    for i in range(1, MANIFEST['total']+1):
        out=OUT / f'.graphify_chunk_{i:02d}.json'
        if out.exists():
            continue
        data=make_chunk(i)
        out.write_text(json.dumps(data, indent=2, ensure_ascii=False), encoding='utf-8')
        made.append((i,len(data['nodes']),len(data['edges']),len(data['hyperedges'])))
    print(f'Generated fallback semantic chunks: {len(made)}')
    for row in made[:20]:
        print('chunk %02d: %d nodes, %d edges, %d hyperedges' % row)
    if len(made)>20:
        print('...')

if __name__ == '__main__':
    main()
