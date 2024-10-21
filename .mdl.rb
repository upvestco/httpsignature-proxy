# Enable all rules by default
all

exclude_rule 'MD024'
exclude_rule 'MD026'

rule 'MD013', :line_length => 160

rule 'MD007', :indent => 2

rule 'MD009', :br_spaces => 2

rule 'MD010', :code_blocks => true
