
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

function onCompetitionSelected(r) {
  console.log("Race selected:")
  console.log(r)

  if( checkError(r) )
    return

  hideAll();

  let title = "id: " + String(r["RaceStatus"]["CompetitionId"]);
  $("#competition_view").show();
  $(".block_hdr_title_value").text(title);
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

