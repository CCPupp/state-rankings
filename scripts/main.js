$(document).ready(function() {
    $("#submitPlayer").on('click', function() {
        var ID = $("#ID").val();
        $.ajax({
            url: "/submitPlayer",
            method: "GET",
            contentType: "application/x-www-form-urlencoded",
            data: {
                ID: ID,
            },
            success: function(data) {
                $("#response").html(data);
            },
        });
    });
});