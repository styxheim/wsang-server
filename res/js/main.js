
var TerminalString = "ad914";

function get_competition_data(competitionId, timestamp, onSuccess) {
  $.getJSON("/api/data/" + String(competitionId) + "/" +
            String(timestamp) + "/" + TerminalString,
            onSuccess);
}

function get_competition_list(onSuccess) {
  $.getJSON("/api/admin/competitions/" + TerminalString, onSuccess)
}

function hideAll() {
  $(".inithide").hide()
}

function checkError(r) {
  if( r.length == 0 ) {
    $("#error_view").show()
    $("#error_text").text("Ответ от сервера нулевой длины")
    return true;
  }
  if( r["Error"] ) {
    $("#error_view").show()
    $("#error_text").text(r["Error"]["Text"])
    return true
  }
  return false
}

function constructRaceStatus() {
  let r = {
        "Gates": [],
        "Penalties": [0],
        "Disciplines": [],
        "CompetitionName": $("#competition_name_view").text(),
        "CompetitionId": Number($("#competition_id_view").text()),
  };

  let gc = $("#gates_edit_store").val();
  $.each(gc.split(","), function(i, el) {
    let n = Number(el);
    if( el.trim() == "" || isNaN(n) )
      return;
    r["Gates"].push(Number(el));
  });

  let pc = $("#penalties_edit_store").val();
  r["Penalties"] = r["Penalties"].concat(num_string_to_array(pc));

  $.each($(".d_edit_name"), function(i, el) {
    let d_id = Number(el.id.split("_")[2]);
    let d = {"Id": d_id,
             "Name": el.textContent,
             "Gates": []
            };

    $.each($(".d_gate_" + String(d_id)), function(i, ga) {
      if( ga.checked ) {
        d["Gates"].push(Number(ga.value));
      }
    });

    r["Disciplines"].push(d);
  });

  return r;
}

function addDiscipline() {
  let r = constructRaceStatus();
  let discipline = {
    "Id": 1,
    "Name": "Новая дисциплина",
    "Gates": []
  };

  let ids = Array.from(r["Disciplines"], x => x["Id"]);
  if( ids.length > 0 ) {
    discipline["Id"] = Math.max.apply(null, ids) + 1;
  }

  console.log(JSON.stringify(r, null, 2));
  r["Disciplines"].push(discipline);
  updateRaceView(r);
}

function delDiscipline(id) {
  let r = constructRaceStatus();
  let disciplines = [];
  id = Number(id);

  for( k in r["Disciplines"] ) {
    if( r["Disciplines"][k]["Id"] == id ) {
      continue;
    }
    disciplines.push(r["Disciplines"][k]);
  }

  r["Disciplines"] = disciplines;
  updateRaceView(r);
}

function updateDisciplines(r) {
  let discipline_edit_container = $("#discipline_edit_container");
  discipline_edit_container.empty();
  try {
    let gates_count = r["Gates"].length
    for( let d in r["Disciplines"] ) {
      let discipline = r["Disciplines"][d];
      let html_discipline_id = "d_edit_" + discipline["Id"];
      let events = "";
      events += " onmouseenter='discipline_name_focus(\"" + html_discipline_id + "\");'";
      events += " onmouseleave='discipline_name_focusout(\"" + html_discipline_id + "\");'";
      let n = $("<tr></tr>")
      let l = $("<th " + events  + " id='" + html_discipline_id + "' class='d_edit_name' colspan='" + String(gates_count) + "'>" + discipline["Name"] + "</th>");
      l.appendTo(n);
      l = $("<th rowspan='2' class='d_del_container'><input type='button' onClick='delDiscipline(\"" + discipline["Id"] + "\")' value='Удалить'/></th>");
      l.appendTo(n);
      n.appendTo(discipline_edit_container);
      n = $("<tr></tr>")
      for( let g in r["Gates"] ) {
        let gate = r["Gates"][g];
        checked = "";
        if( discipline["Gates"].indexOf(gate) > -1 ) {
          checked = "checked='checked' "
        }
        let id = "d_gate_" + String(discipline["Id"]) + "_" + String(gate);
        let discipline_class = "d_gate_" + String(discipline["Id"])
        let td = $("<td></td>");
        checkbox = $("<input class='gate_selector " + discipline_class + "' type='checkbox' " + checked + " id='" + id + "' value='" + String(gate) + "'/>")
        let label = $("<label for='" + id + "'></label>");
        $("head").append("<style class='discipline_gates_style'>#" + id + " + ::after{content: '" + String(gate) + "'}</style>");
        checkbox.appendTo(td);
        label.appendTo(td);
        td.appendTo(n);
      }
      n.appendTo(discipline_edit_container);
    }
  } catch( e ) {
    console.log(e);
  }
}

function updateRaceView(r) {
  let title = "id: " + String(r["CompetitionId"]);
  $(".garbage").hide();
  $(".garbage").blur();
  $(".garbage").remove();
  if( r["CompetitionName"] )
    title = r["CompetitionName"]
  $("#competition_view").show();
  $("#competition_name_view").text(title);
  $("#competition_id_view").text(r["CompetitionId"]);

  console.log(JSON.stringify(r, null, 2));

  let penalties = "";
  try {
  for(let p in r["Penalties"]) {
    if( p == 0 )
      continue;
    penalties += String(r["Penalties"][p]) + ", ";
  }
  } catch( e ) {
    console.log(e);
  }
  $("#penalties_edit_store").val(penalties);
  $("#penalties_edit").val(penalties);

  let gates = "";
  try {
    for(let g in r["Gates"]) {
      gates += String(r["Gates"][g]) + ", ";
    }
  } catch( e ) {
    console.log(e);
  }
  $("#gates_edit_store").val(gates);
  $("#gates_edit").val(gates);

  updateDisciplines(r);
}

function onCompetitionSelected(r) {
  console.log("Race selected:")
  console.log(r)

  if( checkError(r) )
    return;

  hideAll();

  if( r["RaceStatus"] )
    updateRaceView(r["RaceStatus"]);

  let x = $("<pre>" + JSON.stringify(r, null, 2) + "</pre>")
  let c = $("#competition_view_content")
  c.empty()
  x.appendTo(c)
}

function selectCompetition(CompetitionId) {
  get_competition_data(CompetitionId, 0, onCompetitionSelected);
}

function addCompetitionButtom(CompetitionId, CompetitionName, is_active) {
  let title = "id: " + String(CompetitionId);
  let onclick = "onclick='selectCompetition(" + String(CompetitionId) + ")'";
  if( CompetitionName )
    title = CompetitionName;
  let btn_class = ""
  if( is_active )
    btn_class = "active";
  let input = $("<a href='#' class='block_btn " + btn_class + "' id='competition_" + String(CompetitionId) + "'" + onclick + ">" + title + "</a>");
  input.appendTo($("#competition_selector_list"));
}

function onCompetitionList(result) {
  if( checkError(result) )
    return;

  console.log("List acquired: ");
  console.log(result);
  $("#competition_selector_list").empty();
  for(let k in result["Competitions"]) {
    let comp = result["Competitions"][k];
    if( !comp["CompetitionId"] )
      continue;
    addCompetitionButtom(comp["CompetitionId"], comp["CompetitionName"]);
  }
  $("#competition_selector").show();
}

function toPrev(cb) {
  hideAll();

  cb()
}

function main() {
  console.log("Admin initialized")
  get_competition_list(onCompetitionList)
}

function discipline_name_focus(id) {
  let name_container = $("#" + id);
  let name_edit = $("<input id='edit_" + id + "' type='text' />");
  let str = name_edit.val(name_container.text());
  let name_edit_old = $("#edit_" + id);

  name_container.text("");
  if( name_edit_old.length ) {
    name_edit = name_edit_old;
    name_edit.show();
  } else {
    name_edit.appendTo(name_container);
  }

  name_edit.select();
}

function discipline_name_focusout(id) {
  let name_container = $("#" + id);
  let name_edit = $("#edit_" + id);

  name_container.text(name_edit.val());
  name_edit.blur();
  name_edit.hide();
}


function competition_name_focus() {
  let competition_name_edit = $("#competition_name_edit");
  let competition_name_view = $("#competition_name_view");

  competition_name_view.hide();
  competition_name_edit.show();
  competition_name_edit.val(competition_name_view.text());
  competition_name_edit.select();
  competition_name_edit.focus();
}

function competition_name_focusout() {
  let competition_name_edit = $("#competition_name_edit");
  let competition_name_view = $("#competition_name_view");
  let text = competition_name_edit.val();

  competition_name_edit.hide();
  competition_name_view.show();
  if( text ) {
    competition_name_view.text(text);
  }
}

function num_string_to_array(str) {
  let ar = [];

  $.each(str.split(","), function(i, el) {
    let n = Number(el);
    if( el.trim() == "" || isNaN(n) )
      return;
    ar.push(n);
  });

  return ar;
}

function arrays_is_equal(a, b) {
  return (JSON.stringify(a) == JSON.stringify(b));
}

function gp_apply(gp) {
  let edit_id = gp.id;
  let store_id = gp.id + "_store";
  let btn_id = gp.id + "_apply";

  $("#" + store_id).val($(gp).val());
  $(".gp_apply").hide();

  r = constructRaceStatus();
  updateRaceView(r);
}

function gp_keyup(gp) {
  let edit_id = gp.id;
  let store_id = gp.id + "_store";
  let btn_id = gp.id + "_apply";

  let edit = $(gp);
  let store = $("#" + store_id);

  if(!arrays_is_equal(num_string_to_array(edit.val()),
                      num_string_to_array(store.val()))) {
    let btn = $("#" + btn_id);
    btn.click(function() { gp_apply(gp); });
    btn.show();
  } else {
    $("#" + btn_id).hide();
  }
}

function raceUploadEnd() {
  $("#competition_view_submit_btn").prop("disabled", false);
}

function raceUploadFail(v) {
  console.log("POST failed");
  console.log(v);
  checkError({"Error": {"Text": "Неожиданный ответ от сервера. Возможно, сервер недоступен."}});
}

function raceUploadResult(data) {
  console.log("POST result");
  console.log(data);

  if( checkError(data) )
    return;
}

function raceUpload() {
  let r = constructRaceStatus();

  $("#competition_view_submit_btn").prop("disabled", true);

  console.log("POST");
  console.log(r);
  $.post("/api/admin/competition/set/" + TerminalString,
         JSON.stringify(r), raceUploadResult, "json").fail(raceUploadFail).always(raceUploadEnd);
}
