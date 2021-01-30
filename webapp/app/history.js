function showHistory() {
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/dailystats/id/'+id_token);
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   renderHistory(xhr.responseText);
			 };
	  xhr.send();
  }

var map
function renderHistory(raw) {
	if (map==null) {
     		map = $('#world-map').vectorMap({map: 'world_mill',zoomButtons : false, backgroundColor:'white',regionStyle: { initial: { fill: '#dc3545' }, hover: { fill: 'black' } }});
	}
	map.updateSize()
}
