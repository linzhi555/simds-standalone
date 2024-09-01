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
        let timeinfo = "模拟器迭代次数:" + data.updateCount + "     " + "模拟器运行时长:" + data.upTime
        $("#timeLabel").text(timeinfo);
    }
}

$(document).ready(function() {
    $("#netTable").hide()

    $("#filterInput").on("keyup", function() {
        var value = $(this).val().toLowerCase();
        $("#dataTable tr").filter(function() {
            $(this).toggle(
                $(this).children("td").eq(1).text().toLowerCase().indexOf(value) > -1,
            );
        });
    });

    $("#inputBox").on("keydown", function(event) {
        if (event.key === "Enter") {
            getDataFromSever();
        }
    });
    $("#submit").on("click", getDataFromSever);

    $("#toggleDisplay").on("click", toggleDisplay);

    $(".inspectActor").click(function() {
        const name = $(this).closest("tr").find("td").eq(0).text();
        alert("data :" + name);
    });
});


function toggleDisplay() {
    $("#netTable").toggle()
    $("#nodeTable").toggle()
}

