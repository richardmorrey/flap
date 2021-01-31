var historyInit=false
function showHistory() {
	if (!historyInit)
	{
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/flighthistory/id/'+id_token+"/b/5/n/1");
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   renderHistory(xhr.responseText);
			 };
	  xhr.send();
       }
  }
function renderHistory(text) {
	$('#world-map').vectorMap({map: 'world_mill',zoomButtons : false, backgroundColor:'white',regionStyle: { initial: { fill: '#dc3545' }, hover: { fill: 'black' } }});
	$('#world-map').vectorMap('get', 'mapObject').updateSize();
	populateBoard(text)
	historyInit=true
}

function populateBoard(text) {
	$("#cstub").after("<table class='table table-dark table-sm'><thead id='dt'><tr><th scope='col'><b>DATE</b></th><th scope='col' colspan='2'><b>DEPART</b></th><th scope='col' colspan='2'><b>ARRIVE</b></th></tr></thead></tbody></table>")
	rows=""
	raw = JSON.parse(text)
	for (i in raw) {
		for (j in raw[i].Journeys) {
			for (f in raw[i].Journeys[j].Flights) {
				cf = raw[i].Journeys[j].Flights[f]
				start = moment.utc(cf.Start)
				end  = moment.utc(cf.End)
				rows += "<tr><td>" + start.format('YYYY-MM-DD') + "</td><td>" + start.format("hh:mm") + "</td><td>" + cf.From + "</td><td>" + end.format("hh:mm") + "</td><td>" + cf.To + "</td></tr>"
			}
		}
	}
	$('#dt').append(rows)

}
