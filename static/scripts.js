let SUBDOMAIN_PATTERN = /.*\.challonge*/i

var removedPlayers  = []
var newPlayers = []

$(function() {
    $("#newPlayers").hide()
    $(".rmPlayer").hide()
    $(".includePlayer").hide()
    $("#btnTournamentSubmit").hide()

    $("#frmPool").submit(function(event) {
        var name = $("#inpName")
        var players = $("#inpPlayers")

        if(name.val().length <= 0 || players.val().length <= 0) {
            alert("Empty fields")
            event.preventDefault()
        }
    })

    $("#btnAddPlayer").click(function() {
        $("#divPlayers").append(`<div id="divPlayer"><input class="divPlayer" type="text" placeholder="Player Name" name="player"></div>`)
    })

    $("#btnEdit").click(function() {
        $("#btnTournamentSubmit").hide()
        $("#newPlayers").show()
        $(".rmPlayer").show()
        $("#btnEdit").hide()
        $("#btnTournament").show()
        $(".includePlayer").hide ()
    })

    $("#btnTournament").click(function() {
        $(".includePlayer").show()
        $("#newPlayers").hide()
        $(".rmPlayer").hide()
        $("#btnEdit").show()
        $("#btnTournament").hide()
        $("#btnTournamentSubmit").show()

    })

    $("#btnSubmitEdit").click(function() {
        // Take removedPlayers array and newPlayers and submit them

        let players = $(".divPlayer")
        for(var i = 0; i < players.length; i++) {
            let player = $(players[i]).val()
            if (player.length > 0) {
                newPlayers.push(player)
            }
        }
        let poolName = $("#poolTitle").text()

        $.ajax({
            type : "POST",
            contentType : 'application/json',
            url : "/pool/edit/",
            data : JSON.stringify({
                    'name' : poolName,
                    'newP' : newPlayers,
                    'remP' : removedPlayers
                }
            ),
            success : function() {
                window.location.href = "/pool/" + poolName
            },
            error : function(statusCode) {
                console.log("error: " + statusCode)
            }
        })
    })

    $("#btnTournamentSubmit").click(function() {
        var includedPlayers = []
        let checkedPlayers = $('.chkInclude:checked')
        if(checkedPlayers.length == 0) {
            alert("No players selected.")
            return
        }

        for(var i = 0; i < checkedPlayers.length; i++) {
            includedPlayers.push($(checkedPlayers[i]).val())
        }

        let poolName = $("#poolTitle").text()
    $.ajax({
            type : "POST",
            contentType : 'application/json',
            accept: 'application/json',
            url : "/pool/tournament/",
            data : JSON.stringify({
                    'players' : includedPlayers,
                    'name' : poolName
                }
            ),
            success : function(data) {
                let seedResults = $("#seedResults")
                $(seedResults).empty()

                $(seedResults).append("<table><tr><th>Player</th><th>Seed</th></tr>")

                for(var i = 0; i < data.length; i++) {
                    let p = data[i].name
                    let s = i+1
                    $(seedResults).append("<tr><td>" + p + "</td><td>" + s + "</td></tr>")
                }
                $(seedResults.append("</table>"))

            },
            error : function(statusCode) {
                console.log("error: " + statusCode)
            }
        })
    })

    $("#btnChallonge").click(function() {
        let url = $("#challongeURL").val()
        if(url.length === 0) {
            alert("Empty URL")
            return
        }

        let baseURL = "https://api.challonge.com/v1/tournaments/"
        // Assume URL is correct if it exists


        // https://challonge.com/manyulatest
        // ->
        // https://api.challonge.com/v1/tournaments/manyulatest?api_key=hRnTyyGav36S6bxcJzhY37kQktPrHs33mmTUXjti&include_participants=1&include_matches=1

        //https://4qs.challonge.com/4qm127

        var apiURL = ""
        if(SUBDOMAIN_PATTERN.test(url)) {
            var urlAr = url.split("//")
            var subdomain = ""

            // Assume nothing before subdomain (i.e. no "www.")
            for(var i = 0; i  < urlAr[1].length; i++) {
                if(urlAr[1].charAt(i) != '.')
                    subdomain = subdomain + urlAr[1].charAt(i);
                else
                    break
            }

            // Tournament ID
            urlAr = urlAr[1].split("/")
            let tournamentId = urlAr[1]
            apiURL = baseURL + subdomain + "-" + tournamentId

        }
        else {
            let urlAr = url.split("/")
            let tournamentId = urlAr[3]
            apiURL = baseURL + tournamentId
        }

        // Add query parameters
        apiURL = apiURL + "?api_key=hRnTyyGav36S6bxcJzhY37kQktPrHs33mmTUXjti&include_participants=1&include_matches=1"
        let poolName = $("#poolTitle").text()

        $.ajax({
            type : "POST",
            contentType : 'application/json',
            url : "/pool/glicko/",
            data : JSON.stringify({
                    'name' : poolName,
                    'apiURL' : apiURL,
                }
            ),
            success : function() {

                window.location.href = "/pool/" + poolName
            },
            error : function(statusCode) {
                console.log("error: " + statusCode)
            }
        })

    })

    $(".btnRemove").click(function(el) {
        removedPlayers.push($(el.target).attr("data-player"))
        $(el.target).parent().parent().remove()
    })
})