$(document).ready(function() {
  // capture the click and extract the input bxo;
  $("#newpage-submit").on("click", function(e) {
    var newPage = $("#new-page").val()
    newPage = newPage.replaceAll(" ", "-")
    window.location.href = "/view/"+newPage
    return false
  })

  $("#newchildpage-submit").on("click", function(e) {
    var newPage = $("#new-child-page").val()
    newPage = newPage.replaceAll(" ", "-")
    window.location.href = window.location.pathname+"/"+newPage
    return false
  })
})
