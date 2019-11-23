
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

  $.each($("#gates_input").val().split(","), function(i, el) {
    n = Number(el);
    if( n != undefined )
      r["Gates"].push(Number(el));
  });

  $.each($("#penalties_input").val().split(","), function(i, el) {
    n = Number(el);
    if( n != undefined )
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
        d["Gates"].push(ga.value);
      }
    });

    r["Disciplines"].push(d);
  });

  console.log(r);
}

function addDiscipline() {
  constructRaceStatus();
}

function updateDisciplines(r) {
  let discipline_edit_container = $("#discipline_edit_container");
  let gates_count = r["Gates"].length
  discipline_edit_container.empty();
  $(".gates_style").remove();
  for( let d in r["Disciplines"] ) {
    let discipline = r["Disciplines"][d];
    let l = $("<tr><th id='d_edit_" + discipline["Id"] + "' class='d_edit_name' colspan='" + String(gates_count) + "'>" + discipline["Name"] + "</th></tr>");
    l.appendTo(discipline_edit_container);
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
  if( r["CompetitionName"] )
    title = r["CompetitionName"]
  $("#competition_view").show();
  $("#competition_name_view").text(title);
  $("#competition_id_view").text(r["CompetitionId"]);

  let penalties = "";
  for(let p in r["Penalties"]) {
    if( p == 0 )
      continue;
    penalties += String(r["Penalties"][p]) + ", ";
  }

  let gates = "";
  for(let g in r["Gates"]) {
    gates += String(r["Gates"][g]) + ", ";
  }

  $("#penalties_input").val(penalties);
  $("#gates_input").val(gates);
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
