-- The 2023 seed used "Albert Adrià" / "Essència Adrià", a name that does not
-- exist on vicens.com. The real, currently-live name for this solo-chef line
-- is "Adrià Natura" (see docs/design-deliverables Catalog Audit reconciliation).
update "Classes"
set "Name" = 'Adrià Natura',
    "Description" = 'La col·laboració més antiga de la casa: Albert Adrià porta postres icòniques d''elBulli al llenguatge del torró'
where "Id" = '4';

update "Torrons"
set "Name" = replace("Name", '- Albert Adrià', '- Adrià Natura')
where "Class" = '4';
