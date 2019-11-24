
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

  let gc = $("#gates_container").clone().children().remove().end();
  $.each(gc.text().split(","), function(i, el) {
    let n = Number(el);
    if( el.trim() == "" || n == NaN )
      return;
    r["Gates"].push(Number(el));
  });

  let pc = $("#penalties_container").clone().children().remove().end();
  $.each(pc.text().split(","), function(i, el) {
    let n = Number(el);
    if( el.trim() == "" || n == NaN )
      return;
    r["Penalties"].push(Number(el));
  });

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
  let gates_count = r["Gates"].length
  discipline_edit_container.empty();
  $(".gates_style").blur();
  $(".gates_style").hide();
  $(".gates_style").remove();
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

  if( r["Penalties"].length > 1 ) {
    let penalties = "";
    for(let p in r["Penalties"]) {
      if( p == 0 )
        continue;
      penalties += String(r["Penalties"][p]) + ", ";
    }

    $("#penalties_container").text(penalties);
  } else {
    let x = $("<span class='garbage'>без штрафов</span>");
    x.appendTo($("#penalties_container"));
  }

  if( r["Gates"].length > 0 ) {
    let gates = "";
    for(let g in r["Gates"]) {
      gates += String(r["Gates"][g]) + ", ";
    }
    $("#gates_container").text(gates);
  } else {
    let x = $("<span class='garbage'>без ворот</span>");
    x.appendTo($("#gates_container"));
  }

  updateDisciplines(r);
}

function onCompetitionSelected(r) {
  console.log("Race selected:")
  console.log(r)

  if( checkError(r) )
    return

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
  let name_edit = $("<input id='edit_" + id + "' type='text' value='" + name_container.text() + "'/>");

  name_container.text("");
  name_edit.appendTo(name_container);
  name_edit.select();
}

function discipline_name_focusout(id) {
  let name_container = $("#" + id);
  let name_edit = $("#edit_" + id);

  name_container.text(name_edit.val());
  name_edit.blur();
  name_edit.hide();
  name_edit.remove();
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

function gp_edit_focus(gp_container)
{
  gp_container = $(gp_container);
  let edit = $("<input class='garbage' id='edit_" + gp_container.attr("id") + "' type='text' value=''/>");
  let str = gp_container.text();

  gp_container.text("");
  edit.appendTo(gp_container);
  edit.focus();
  edit.val(str);
  if( str.length == 0 ) {
    edit.width(20 * 8);
  } else {
    edit.width(str.length * 8);
  }
}

function penalties_edit_focusout(p_container) {
  p_container = $(p_container);
  let edit = $("#edit_" + p_container.attr("id"));
  let str = edit.val();
  let a = Array.from(str.split(","));
  let penalties = [0];

  edit.blur();
  edit.hide();
  edit.remove();

  $.each(a, function( i, v ) {
    if( v.trim() === "" || isNaN((v = Number(v))) ) {
      return;
    }
    penalties.push(v);
  });

  console.log(penalties);
  str = penalties.join(", ");

  let r = constructRaceStatus();
  r["Penalties"] = penalties;
  updateRaceView(r);
}


function gates_edit_focusout(gates_container) {
  gates_container = $(gates_container);
  let edit = $("#edit_" + gates_container.attr("id"));
  let str = edit.val();
  let a = Array.from(str.split(","));
  let gates = [];

  edit.blur();
  edit.hide();
  edit.remove();
  $.each(a, function( i, v ) {
    if( v.trim() === "" || isNaN((v = Number(v))) )
      return;
    if( gates.indexOf(v) > -1 )
      return;
    gates.push(v);
  });

  str = gates.join(", ");

  let r = constructRaceStatus();

  r["Gates"] = gates;
  updateRaceView(r);
}
