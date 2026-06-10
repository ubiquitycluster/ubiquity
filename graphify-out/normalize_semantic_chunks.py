import json
import re
from pathlib import Path

ROOT = Path('/Users/ccoates/Documents/ubiquity')
OUT = ROOT / 'graphify-out'
VALID_FILE_TYPES = {'code','document','paper','image','rationale','concept'}
VALID_EDGE_RELATIONS = {'calls','implements','references','cites','conceptually_related_to','shares_data_with','semantically_similar_to','rationale_for'}
VALID_CONF = {'EXTRACTED','INFERRED','AMBIGUOUS'}

def clean_id(s):
    s = str(s or 'node').lower()
    s = re.sub(r'[^a-z0-9]+', '_', s).strip('_')
    return s or 'node'

def infer_file_type(node):
    ft = node.get('file_type') or node.get('type') or node.get('node_type')
    if ft in VALID_FILE_TYPES:
        return ft
    sf = str(node.get('source_file') or '')
    ext = Path(sf).suffix.lower()
    if ext in {'.go','.py','.sh','.bash','.tf','.json'}:
        return 'code'
    if ext in {'.png','.jpg','.jpeg','.gif','.webp','.svg'}:
        return 'image'
    # category-ish old types become concept
    old = str(ft or '').lower()
    if old in {'chart','service','pipeline','workflow','secret','goal','plan','resource','component','concept'}:
        return 'concept'
    return 'document' if sf else 'concept'

def infer_label(old_id, node):
    if node.get('label'):
        return str(node['label'])
    raw = str(old_id or node.get('id') or 'Node')
    raw = raw.split(':')[-1]
    return re.sub(r'[_-]+',' ',raw).strip().title() or 'Node'

def normalize_chunk(path):
    data=json.loads(path.read_text(encoding='utf-8'))
    idmap={}
    nodes=[]; seen=set()
    for n in data.get('nodes', []):
        old=n.get('id') or n.get('label') or 'node'
        nid=clean_id(old)
        base=nid; c=2
        while nid in seen:
            nid=f'{base}_{c}'; c+=1
        seen.add(nid); idmap[str(old)]=nid
        sf=n.get('source_file') or n.get('file') or n.get('path') or None
        nodes.append({
            'id': nid,
            'label': infer_label(old,n),
            'file_type': infer_file_type(n),
            'source_file': sf,
            'source_location': n.get('source_location'),
            'source_url': n.get('source_url'),
            'captured_at': n.get('captured_at'),
            'author': n.get('author'),
            'contributor': n.get('contributor'),
        })
    node_ids={n['id'] for n in nodes}
    edges=[]; edge_seen=set()
    for e in data.get('edges', []):
        src=idmap.get(str(e.get('source')), clean_id(e.get('source')))
        tgt=idmap.get(str(e.get('target')), clean_id(e.get('target')))
        if src not in node_ids or tgt not in node_ids:
            continue
        rel=e.get('relation') or 'conceptually_related_to'
        if rel not in VALID_EDGE_RELATIONS:
            rel='conceptually_related_to'
        conf=e.get('confidence') or 'INFERRED'
        if conf not in VALID_CONF:
            conf='INFERRED'
        score=e.get('confidence_score')
        if not isinstance(score, (int,float)):
            score = 1.0 if conf == 'EXTRACTED' else (0.2 if conf == 'AMBIGUOUS' else 0.75)
        if conf == 'EXTRACTED': score=1.0
        elif conf == 'AMBIGUOUS': score=min(max(float(score),0.1),0.3)
        else:
            allowed=[0.95,0.85,0.75,0.65,0.55]
            score=min(allowed, key=lambda x: abs(x-float(score)))
        key=(src,tgt,rel,e.get('source_file'))
        if key in edge_seen: continue
        edge_seen.add(key)
        edges.append({
            'source': src,
            'target': tgt,
            'relation': rel,
            'confidence': conf,
            'confidence_score': score,
            'source_file': e.get('source_file'),
            'source_location': e.get('source_location'),
            'weight': e.get('weight', 1.0),
        })
    hypers=[]
    for h in data.get('hyperedges', [])[:3]:
        hs=[]
        for x in h.get('nodes', []):
            nx=idmap.get(str(x), clean_id(x))
            if nx in node_ids and nx not in hs:
                hs.append(nx)
        if len(hs) < 3:
            continue
        conf=h.get('confidence') if h.get('confidence') in {'EXTRACTED','INFERRED'} else 'INFERRED'
        score=h.get('confidence_score')
        if not isinstance(score,(int,float)): score=0.75
        hypers.append({
            'id': clean_id(h.get('id') or h.get('label') or 'hyperedge'),
            'label': h.get('label') or infer_label(h.get('id'), h),
            'nodes': hs,
            'relation': h.get('relation') if h.get('relation') in {'participate_in','implement','form'} else 'participate_in',
            'confidence': conf,
            'confidence_score': float(score),
            'source_file': h.get('source_file') or (nodes[0]['source_file'] if nodes else None),
        })
    fixed={'nodes':nodes,'edges':edges,'hyperedges':hypers,'input_tokens':int(data.get('input_tokens') or 0),'output_tokens':int(data.get('output_tokens') or 0)}
    path.write_text(json.dumps(fixed, indent=2, ensure_ascii=False), encoding='utf-8')
    return len(nodes), len(edges), len(hypers)

def main():
    for p in sorted(OUT.glob('.graphify_chunk_*.json')):
        n,e,h=normalize_chunk(p)
        print(f'{p.name}: {n} nodes {e} edges {h} hyperedges')

if __name__ == '__main__':
    main()
