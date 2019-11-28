
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

function constructCompetition() {
  return {"RaceStatus": constructRaceStatus(),
          "TerminalStatus": constructTerminalStatus()};
}

function constructTerminalStatus() {
  return []
}

function constructRaceStatus() {
  let r = {
        "Gates": [],
        "Penalties": [0],
        "Disciplines": [],
        "CompetitionName": $("#competition_name_edit").val(),
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
  // TODO: update terminal view
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
  // TODO: update terminal view
  updateRaceView(r);
}

function updateTerminalView(terminals, disps) {
  let terminal_container = $("#terminal_container");
  terminal_container.empty();

  console.log("Update Terminal List");

  for( let k in terminals ) {
    v = terminals[k];
    console.log(v);
    term = $("#terminal_container_tpl").clone();
    term.attr("id", "terminal_container_" + String(v["TerminalId"]));
    term.find("#terminal_id_view").text(v["TerminalId"]);
    term.appendTo(terminal_container);
    term.show();
  }
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

function addRace() {
  let r = {
    "CompetitionId": 0,
    "CompetitionName": "",
    "Gates": [1, 2, 3, 4],
    "Penalties": [0],
    "Discplines": [
      {
        "Id": 1,
        "Name": "Слалом",
        "Gates": [1, 3]
      }
    ]
  };

  hideAll();
  updateCompetitionView({"RaceStatus": r, "TerminalStatus": []}, function() {
    uploadCompetition().done(function() {
      hideAll();
      get_competition_list(onCompetitionList);
    });
  });
}

function updateRaceView(r) {
  let title = "id: " + String(r["CompetitionId"]);
  $("#competition_view").show();
  $("#competition_name_edit").val(r["CompetitionName"]);
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

function updateCompetitionView(c, onsubmit) {

  if( c["RaceStatus"] ) {
    updateRaceView(c["RaceStatus"]);
  }

  if( c["TerminalStatus"] ) {
    let disps = [];
    if( c["RaceStatus"] && c["RaceStatus"]["Disciplines"] ) {
      disps = c["RaceStatus"]["Disciplines"];
    }
    updateTerminalView(c["TerminalStatus"], disps);
  }

  if( onsubmit ) {
    $("#competition_view_submit_btn").off("click");
    $("#competition_view_submit_btn").on("click", onsubmit);
  }
}

function onCompetitionSelected(c) {
  console.log("Race selected:")
  console.log(c)

  if( checkError(c) )
    return;

  hideAll();

  updateCompetitionView(c, uploadCompetition);
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
  result["Competitions"].sort(function(a, b) {
    if( a["CompetitionId"] > b["CompetitionId"] ) return 1;
    if( a["CompetitionId"] < b["CompetitionId"] ) return -1;
    return 0;
  });
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
    btn.off("click");
    btn.on("click", function() { gp_apply(gp); });
    btn.show();
  } else {
    $("#" + btn_id).hide();
  }
}

function competitionUploadEnd() {
  $("#competition_view_submit_btn").prop("disabled", false);
}

function competitionUploadFail(v) {
  console.log("POST failed");
  console.log(v);
  checkError({"Error": {"Text": "Неожиданный ответ от сервера. Возможно, сервер недоступен."}});
}

function competitionUploadResult(data) {
  console.log("POST result");
  console.log(data);

  if( checkError(data) )
    return;
}

function uploadCompetition() {
  let c = constructCompetition();

  $("#competition_view_submit_btn").prop("disabled", true);

  console.log("POST");
  console.log(c);

  return $.post("/api/admin/competition/set/" + c["RaceStatus"]["CompetitionId"]  + "/" + TerminalString,
                JSON.stringify(c), competitionUploadResult, "json").fail(competitionUploadFail).always(competitionUploadEnd);
}
