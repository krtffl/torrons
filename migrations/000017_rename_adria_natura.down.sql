update "Torrons"
set "Name" = replace("Name", '- Adrià Natura', '- Albert Adrià')
where "Class" = '4';

update "Classes"
set "Name" = 'Albert Adrià',
    "Description" = 'Essència Adrià'
where "Id" = '4';
