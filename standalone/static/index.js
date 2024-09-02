var lastResponse

function getDataFromSever() {
    var userInput = $("#inputBox").val();
    const url = `http://localhost:8079/run?cmd=${userInput}`;
    $.ajax({
        url: url,
        type: "GET",
        dataType: 'json',
        success: function(response) {
            dealResponse(response)
        },
        error: function(xhr, status, error) {
            console.error("错误:", xhr, status, error);
            $("#outputText").text("获取数据失败：" + error);
        },
    });
}

function dealResponse(data) {
    if (data.error != "null") {
        $("#outputText").text(data.error);
    } else {
        $("#updateCount").text(data.updateCount)
        $("#upTime").text(data.upTime)
        updateNodesTable(data.nodesState)
        updateNetTable(data.netState)
        doFilter()
    }
}

function updateNodesTable(data) {
    console.log(data)
    console.log(data.length)
    console.log(Array.isArray(data))

    $("#nodesTable tbody").empty()
    data.forEach(actor => {
        let lastmsg = "null"
        let curmsg = "null"
        if (actor.isBusy == "true") {
            lastmsg = actor.msg
        } else {
            curmsg = actor.msg
        }


        let newrow = `
        <tr>
            <td>  ${actor.name} </td>
            <td>  ${actor.node} </td>
            <td>  ${actor.isBusy} </td>
            <td>  ${actor.difficulty} </td>
            <td>  ${actor.progress} </td>
            <td>  ${lastmsg} </td>
            <td>  ${curmsg} </td>
        </tr>
        `
        $("#nodesTable tbody").append(newrow)
    })
}

function updateNetTable(data) {
    console.log(data.waittings)
    $("#netTable tbody").empty()
    if (Array.isArray(data.waittings)) {
        data.waittings.forEach(msg => {
            let newrow = `
        <tr>
            <td>  waitting </td>
            <td>  ${msg.from} </td>
            <td>  ${msg.to} </td>
            <td>  ${msg.head} </td>
            <td>  ${msg.body} </td>
            <td>  ${msg.leftTime} </td>
        </tr>
        `
            console.log("!!!!!!q")
            $("#netTable tbody").append(newrow)
        })
    }

}

$(document).ready(function() {
    $("#netTable").hide()

    $("#inputBox").on("keydown", function(event) {
        if (event.key === "Enter") {
            getDataFromSever();
        }
    });
    $("#submit").on("click", getDataFromSever);

    $("#toggleDisplay").on("click", toggleDisplay);

    $('#filter').on('keyup', doFilter);

});

function doFilter() {
    const filterValue = $("#filter").val().toLowerCase();
    let _dofilt = function(tablename) {
        $(`#${tablename} tbody tr`).filter(function() {
            let rowVisible = false;
            $(this).find('td').each(function() {
                if ($(this).text().toLowerCase().indexOf(filterValue) > -1) {
                    rowVisible = true;
                }
            });
            $(this).toggle(rowVisible);
        });
    }

    _dofilt("nodesTable")
    _dofilt("netTable")
}


function toggleDisplay() {
    $("#netTable").toggle()
    $("#nodesTable").toggle()
}
