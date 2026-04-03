import re

def to_camel_case(s):
    if s == "ID": return "id"
    if s == "Ici1": return "ici1"
    if s == "Ici2": return "ici2"
    if s == "Ici3": return "ici3"
    if s == "Ici4": return "ici4"
    if s == "Ici5": return "ici5"
    if s == "IupOp": return "iupOp"
    if not s: return s
    return s[0].lower() + s[1:]

with open(r'd:\programming\go\sinkronisasi_db\internal\model\models.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    m = re.match(r'^(\t)([A-Z][a-zA-Z0-9_]*)([\t\s]+[\w\.\*]+[\t\s]+`gorm:")([^"]+)(".*)$', line)
    if m:
        field_name = m.group(2)
        gorm_tag = m.group(4)
        if "column:" not in gorm_tag:
            camel = to_camel_case(field_name)
            new_tag = f"column:{camel};{gorm_tag}"
            line = f"{m.group(1)}{field_name}{m.group(3)}{new_tag}{m.group(5)}\n"
            
    new_lines.append(line)

with open(r'd:\programming\go\sinkronisasi_db\internal\model\models.go', 'w', encoding='utf-8') as f:
    f.writelines(new_lines)
