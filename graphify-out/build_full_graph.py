import json
from pathlib import Path
import networkx as nx
from graphify.build import build, validate_extraction

ROOT = Path('/Users/ccoates/Documents/ubiquity')
OUT = ROOT / 'graphify-out'

extractions = []
# AST extraction first
ast_path = OUT / '.graphify_ast.json'
if ast_path.exists():
    extractions.append(json.loads(ast_path.read_text(encoding='utf-8')))

errors = []
for p in sorted(OUT.glob('.graphify_chunk_*.json')):
    if not p.name.endswith('.json'):
        continue
    data = json.loads(p.read_text(encoding='utf-8'))
    for k in ['nodes','edges','hyperedges','input_tokens','output_tokens']:
        if k not in data:
            errors.append(f'{p.name}: missing {k}')
    errs = validate_extraction(data)
    if errs:
        errors.extend(f'{p.name}: {e}' for e in errs[:20])
    extractions.append(data)

if errors:
    raise SystemExit('Validation failed:\n' + '\n'.join(errors[:200]))

G = build(extractions, directed=True, dedup=True, root=ROOT)
# Attach metadata expected by graphify tooling.
G.graph['directed'] = True
G.graph['root'] = str(ROOT)
G.graph['extraction_count'] = len(extractions)
G.graph['input_tokens'] = sum(int(e.get('input_tokens') or 0) for e in extractions)
G.graph['output_tokens'] = sum(int(e.get('output_tokens') or 0) for e in extractions)
G.graph['token_cost_note'] = 'AST extraction plus Codex semantic chunks; fallback chunks use deterministic extraction with token counts set to 0.'

data = nx.node_link_data(G, edges='links')
(OUT / 'graph.json').write_text(json.dumps(data, indent=2, ensure_ascii=False), encoding='utf-8')
print(f'Wrote graphify-out/graph.json: {G.number_of_nodes()} nodes, {G.number_of_edges()} edges, extractions={len(extractions)}')
print(f"Token accounting: input_tokens={G.graph['input_tokens']} output_tokens={G.graph['output_tokens']}")
