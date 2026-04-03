import re

def to_camel(s):
    if s in ["id", "version", "status", "type", "gar", "ts", "ash", "tm", "im", "fc", "nar", "adb", "hba"]: return s
    parts = s.split('_')
    return parts[0] + ''.join(word.capitalize() for word in parts[1:])

# Process entity_registry.go
with open(r'd:\programming\go\sinkronisasi_db\internal\service\entity_registry.go', 'r', encoding='utf-8') as f:
    content = f.read()

def repl(m):
    key = m.group(1)
    if '_' in key:
        return f'"{to_camel(key)}":'
    return m.group(0)

# Replace all map string keys that look like "snake_case": 
content = re.sub(r'"([a-z_]+)"\s*:', repl, content)

# specific fix for "result_gar", etc if matched
with open(r'd:\programming\go\sinkronisasi_db\internal\service\entity_registry.go', 'w', encoding='utf-8') as f:
    f.write(content)

# Process generic_repo.go
with open(r'd:\programming\go\sinkronisasi_db\internal\repository\generic_repo.go', 'r', encoding='utf-8') as f:
    repo_content = f.read()

# Fix queries
repo_content = repo_content.replace('"is_deleted = ?"', '"\\"isDeleted\\" = ?"')
repo_content = repo_content.replace('"created_at DESC"', '"\\"createdAt\\" DESC"')

# Fix map assignments like "is_deleted": true -> "isDeleted": true
repo_content = re.sub(r'"([a-z_]+)"\s*:', repl, repo_content)

with open(r'd:\programming\go\sinkronisasi_db\internal\repository\generic_repo.go', 'w', encoding='utf-8') as f:
    f.write(repo_content)

print("Done")
