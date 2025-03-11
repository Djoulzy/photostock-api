refresh materialized view models with data;

update galleries gal set model_name = (
	select mod.model_name from models mod where source_name ilike concat('%', mod.model_name, '%')
)