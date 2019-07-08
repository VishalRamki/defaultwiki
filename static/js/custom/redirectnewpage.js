$(document).ready(function() {
  // capture the click and extract the input bxo;
  $("#newpage-submit").on("click", function(e) {
    var newPage = $("#new-page").val()
    window.location.href = "/view/"+newPage
    return false
  })
})
