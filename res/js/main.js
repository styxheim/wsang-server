
const TerminalString = "ad914";

const START_GATE = -2;
const FINISH_GATE = -3;

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
  let rstat = constructRaceStatus();
  return {"RaceStatus": rstat,
          "TerminalStatus": constructTerminalStatus(rstat["Disciplines"])};
}

function constructTerminalStatus(disps) {
  let terminal_status = [];
  $.each($(".terminal_container"), function(i, v) {
    let term = {};
    v = $(v);
    term["TerminalId"] = v.find("#terminal_id_view").text();
    if( term["TerminalId"] == "" ) {
      return;
    }
    // permissions
    let perms = {};
    $.each(v.find(".terminal_perm"), function(ii, vv) {
      perms[vv.name] = vv.checked;
    });
    term["Permissions"] = perms;
    // disciplines
    let disciplines = [];
    $.each(v.find(".t_d_view"), function(iii, vvv) {
      vvv = $(vvv);
      let disp_gates = [];
      let disp_id = Number(vvv.find(".t_d_id").text());
      let discipline = {"Id": disp_id, "Gates": []};
      $.each(disps, function(x, xx) {
        if( xx["Id"] == disp_id ) {
          if( xx["Gates"] ) {
            disp_gates = xx["Gates"];
          }
          return;
        }
      });
      $.each(vvv.find(".t_d_g_check"), function(iiii, vvvv) {
        if( vvvv.checked ) {
          let gate_id = Number(vvvv.value);
          if( disp_gates.indexOf(gate_id) == -1 &&
              gate_id != START_GATE && gate_id != FINISH_GATE ) {
            return;
          }
          discipline["Gates"].push(gate_id);
        }
      });
      disciplines.push(discipline);
    });
    term["Disciplines"] = disciplines;

    terminal_status.push(term);
  });
  console.log(JSON.stringify(terminal_status));
  return terminal_status;
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
        if( r["Gates"].indexOf(Number(ga.value)) == -1 )
          return;
        d["Gates"].push(Number(ga.value));
      }
    });

    r["Disciplines"].push(d);
  });

  return r;
}

function addDiscipline() {
  let c = constructCompetition();
  let discipline = {
    "Id": 1,
    "Name": "Новая дисциплина",
    "Gates": []
  };

  let ids = Array.from(c["RaceStatus"]["Disciplines"], x => x["Id"]);
  if( ids.length > 0 ) {
    discipline["Id"] = Math.max.apply(null, ids) + 1;
  }

  c["RaceStatus"]["Disciplines"].push(discipline);
  updateCompetitionView(c, null);
}

function delDiscipline(id) {
  let c = constructCompetition();
  let disciplines = [];
  id = Number(id);

  for( k in c["RaceStatus"]["Disciplines"] ) {
    if( c["RaceStatus"]["Disciplines"][k]["Id"] == id ) {
      continue;
    }
    disciplines.push(c["RaceStatus"]["Disciplines"][k]);
  }

  c["RaceStatus"]["Disciplines"] = disciplines;
  updateCompetitionView(c, null);
}

function updateTerminalView(terminals, disps) {
  let terminal_container = $("#terminal_container");
  terminal_container.empty();
  $(".terminal_d_g_class").remove();

  console.log("Update Terminal List");

  for( let k in terminals ) {
    let v = terminals[k];
    if( v["TerminalId"] == TerminalString )
      continue;
    if( !v["Permissions"] ) {
      v["Permissions"] = {"Read": false, "Write": false, "Admin": false};
    }
    term = $("#terminal_container_tpl").clone();
    term.attr("id", "terminal_container_" + String(v["TerminalId"]));
    term.find("#terminal_id_view").text(v["TerminalId"]);
    term.find(".terminal_remove").on("click", function() {
      let c = constructCompetition();
      let new_terminals = []

      for( let m in c["TerminalStatus"] ) {
        if( c["TerminalStatus"][m]["TerminalId"] != v["TerminalId"] ) {
          new_terminals.push(c["TerminalStatus"][m]);
        }
      }
      c["TerminalStatus"] = new_terminals;
      updateCompetitionView(c);
    });
    // Permissions
    term.find("#t_perm_read").prop("id", "t_perm_read_" + v["TerminalId"]).prop("checked", v["Permissions"]["Read"] == true);
    term.find("label[for='t_perm_read']").prop("for", "t_perm_read_" + v["TerminalId"]);
    term.find("#t_perm_write").prop("id", "t_perm_write_" + v["TerminalId"]).prop("checked", v["Permissions"]["Write"] == true);
    term.find("label[for='t_perm_write']").prop("for", "t_perm_write_" + v["TerminalId"]);
    term.find("#t_perm_admin").prop("id", "t_perm_admin_" + v["TerminalId"]).prop("checked", v["Permissions"]["Admin"] == true);
    term.find("label[for='t_perm_admin']").prop("for", "t_perm_admin_" + v["TerminalId"]);
    // Disciplines
    let disp_tpl = term.find("#terminal_discipline_tpl").clone();
    term.find(".terminal_discipline").empty();
    for( k in disps ) {
      let terminal_disp_gates = [];
      let d = disps[k];
      for( let vv in v["Disciplines"] ) {
        if( v["Disciplines"][vv]["Id"] == d["Id"] ) {
          if( v["Disciplines"][vv]["Gates"] ) {
            terminal_disp_gates = v["Disciplines"][vv]["Gates"];
          }
          break;
        }
      }
      disp = disp_tpl.clone();
      disp.prop("id", "terminal_discipline_" + v["TerminalId"] + "_" + d["Id"]);
      disp.find(".t_d_id").text(d["Id"]);
      disp.find(".t_d_name").text(d["Name"]);
      let gate_tpl = disp.find(".t_d_g").clone();
      disp.find(".t_d_container").empty();
      if( !d["Gates"] )
        d["Gates"] = [];
      let list_gates = [START_GATE].concat(d["Gates"].concat([FINISH_GATE]));
      for( k in list_gates) {
        let gate_id = list_gates[k];
        let gate_new = gate_tpl.clone();
        let gate_title = String(gate_id);
        let gate_checked = terminal_disp_gates.indexOf(gate_id) != -1;
        if( gate_id == START_GATE ) gate_title = "S";
        if( gate_id == FINISH_GATE ) gate_title = "F";
        id = "t_d_g_" + v["TerminalId"] + "_" + d["Id"] + "_" + gate_title;
        gate_new.find(".t_d_g_check").prop("id", id).val(gate_id).prop("checked", gate_checked);
        gate_new.find(".t_d_g_label").prop("for", id);
        $("head").append($("<style class='terminal_d_g_class'>#" + id + " + ::after{content: '" + gate_title + "'; }</style>"));
        gate_new.appendTo(disp.find(".t_d_container"));
      }
      disp.show();
      disp.appendTo(term.find(".terminal_discipline"));
    }
    term.appendTo(terminal_container);
    term.show();
  }
}

function updateDisciplines(r) {
  let discipline_edit_container = $("#discipline_edit_container");
  discipline_edit_container.empty();
  $(".discipline_gates_style").remove();
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

  $(".gate_selector").on("click", function() {
    let c = constructCompetition();
    updateCompetitionView(c);
  });
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


function timeDifference(current, previous) {
  var msPerMinute = 60 * 1000;
  var msPerHour = msPerMinute * 60;
  var msPerDay = msPerHour * 24;
  var msPerMonth = msPerDay * 30;
  var msPerYear = msPerDay * 365;

  var elapsed = current - previous;

  if (elapsed < msPerMinute) {
     return Math.round(elapsed/1000) + ' seconds ago';
  } else if (elapsed < msPerHour) {
     return Math.round(elapsed/msPerMinute) + ' minutes ago';
  } else if (elapsed < msPerDay ) {
     return Math.round(elapsed/msPerHour ) + ' hours ago';
  } else if (elapsed < msPerMonth) {
    return 'approximately ' + Math.round(elapsed/msPerDay) + ' days ago';
  } else if (elapsed < msPerYear) {
    return 'approximately ' + Math.round(elapsed/msPerMonth) + ' months ago';
  } else {
    return 'approximately ' + Math.round(elapsed/msPerYear ) + ' years ago';
  }
}

function updateActivityView(a) {
  let container = $("#terminal_activity_container").empty();
  let activity_tpl = $("#terminal_activity_tpl").clone();
  let c = constructCompetition();
  let terminals = [];

  for( let k in c["TerminalStatus"] ) {
    terminals.push(c["TerminalStatus"][k]["TerminalId"]);
  }

  a = a.sort(function(a, b) {
    if( a["Activity"]["LastActivity"] < b["Activity"]["LastActivity"] )
      return 1;
    if( a["Activity"]["LastActivity"] > b["Activity"]["LastActivity"] )
      return -1;
    return 0;
  });

  $.each(a, function(i, activity) {
    if( activity["TerminalId"] == TerminalString ) {
      // hide current terminal
      return;
    }
    if( terminals.indexOf(activity["TerminalId"]) != -1 ) {
      // hide already added terminals
      return;
    }
    let time = timeDifference(new Date(), new Date(activity["Activity"]["LastActivity"]));
    new_activity = activity_tpl.clone();
    new_activity.prop("id", "t_a_v_" + activity["TerminalId"]);
    new_activity.find(".t_a_id").text(activity["TerminalId"]);
    new_activity.find(".t_a_time").text(time);
    new_activity.find(":input[type='button']").on('click', function() {
      let c = constructCompetition();
      activity["Permissions"] = {"Read": true, "Write": true};
      c["TerminalStatus"].push(activity);
      updateCompetitionView(c);
      updateActivityView(a);
    });
    new_activity.appendTo(container);
  });
}

function onActivitySuccess(data) {
  let x = $("<pre>" + JSON.stringify(data) + "</pre>");

  updateActivityView(data["TerminalStatus"]);
}

function get_activities() {
  $.getJSON("/api/admin/activity/" + TerminalString, onActivitySuccess);
}

function main() {
  console.log("Admin initialized")
  setInterval(get_activities, 3000)
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

  c = constructCompetition();
  updateCompetitionView(c, null);
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
  $("#save_view").hide();
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
  $("#save_view").show();

  console.log("POST");
  console.log(c);
  console.log(JSON.stringify(c, null, 2));

  return $.post("/api/admin/competition/set/" + c["RaceStatus"]["CompetitionId"]  + "/" + TerminalString,
                JSON.stringify(c), competitionUploadResult, "json").fail(competitionUploadFail).always(competitionUploadEnd);
}
