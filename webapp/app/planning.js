var  planningInit=false
function showPlanning() {
	if (!planningInit)
	{
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/flighthistory/id/'+id_token+"/b/5/n/1");
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   renderPlanning(xhr.responseText);
			 };
	  xhr.send();
       } else {
       		navbarActive('planning')
       }
  }

function renderPlanning(text) {
	new Calendar('#calendar')
	navbarActive('planning');
	planningInit=true
}

