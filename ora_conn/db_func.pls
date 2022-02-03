create table WAV_CONCAT_HIST(	
    id      number, 
	form    varchar2(4000 byte), 
	fname   varchar2(400 byte), 
	status  number
);
/

create or replace function TEST_FUNC (
    action in varchar2
) return varchar2 
as
    lj_input_data   json_object_t := json_object_t();
    lc_action       varchar2(30);
    lc_result       json_object_t := json_object_t();
    ln_id           integer;
    lc_formula      varchar2(500);
    lc_fname        varchar2(500); 
    ln_status       integer;
    
begin
    lj_input_data := json_object_t.parse(action);
    lc_action     := lower(lj_input_data.get_string('action'));
    
    if lc_action = 'get_id' then
        select max(id)+1 into ln_id from TMP_AAA;
        if ln_id is null then
            ln_id := 1000;
        end if;
    
        lc_result.put('res', 'OK');
        lc_result.put('value', ln_id);  

    elsif lc_action = 'save' then
        ln_id      := lj_input_data.get_number('id');
        lc_formula := lower(lj_input_data.get_string('formula'));
        lc_fname   := lower(lj_input_data.get_string('fname'));
        ln_status  := lj_input_data.get_number('status');
        
        insert into WAV_CONCAT_HIST (
            id, form, fname, status
        ) values (
            ln_id, lc_formula, lc_fname, ln_status
        );
        commit;     
        
        lc_result.put('res', 'OK');

    else
        lc_result.put('res', 'ERROR');  

    end if;    
   
    return lc_result.to_string();
end;
/
