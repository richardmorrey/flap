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
	$('#world-map').vectorMap({map: 'world_mill',zoomButtons : false, backgroundColor:'white',regionStyle: { initial: { fill: '#dc3545' }, hover: { fill: 'black' }} , 
					markerStyle:{initial: {fill: 'black',stroke: 'white',"fill-opacity": 1,"stroke-width": 1,"stroke-opacity": 1,r: 5}}});
	$('#world-map').vectorMap('get', 'mapObject').updateSize();
	populateBoard(text)
        $('#flightList a').on('click', function (e) {
	   e.preventDefault()
           showFlightPath(parseInt($(this).attr("href"),10))
	})
	historyInit=true
}

var flights=[]
function populateBoard(text) {
	var raw = JSON.parse(text)
	var active=" active "
	var tf=0
	var rows=""
	for (i in raw)  {
		for (j in raw[i].Journeys) {
			for (f in raw[i].Journeys[j].Flights) {
				if (tf%5==0) {
					rows+="<div id='flightList' class='list-group carousel-item w-100" + active + "h5'>"
					active=" "
				}
				cf = raw[i].Journeys[j].Flights[f]
				flights.push(cf)
				start = moment.utc(cf.Start)
				end  = moment.utc(cf.End)
				rows += "<a class='list-group-item list-group-item-action text-white bg-dark' href='" + (flights.length-1).toString() + "'>&nbsp;&nbsp;" + start.format('YYYY-MM-DD') +  "&nbsp;&nbsp;" + start.format("hh:mm") + " " + gAirports[cf.From].Iata + "&nbsp;&nbsp;" + end.format("hh:mm") + " " + gAirports[cf.To].Iata + "</a>"
				tf++
				if (tf==5) {
					rows+="</div>"
					tf=0
				}
			}
		}
	}
	if (tf != 5) {
		for (;tf < 5; tf ++) {
			rows += "<a class='list-group-item text-white bg-dark' href='#'>&nbsp;</a>"
		}
		rows+="</div>"
	}

	$('#cstub').append(rows)
}

function showFlightPath(i) {
	cf = flights[i]
	nDots= Math.min(cf.Distance/100,10)
	from = [gAirports[cf.From].Lat, gAirports[cf.From].Lng]
	to = [gAirports[cf.To].Lat, gAirports[cf.To].Lng]
	var latdelta=(to[0] - from[0])/(nDots-1)
	var lngdelta=(to[1] - from[1])/(nDots-1)
	var markers = []
	var lat=from[0]
	var lng=from[1]
	for (var i=0;i < nDots; i++) {
		markers.push({latLng:[lat,lng]})
		lat += latdelta
		lng += lngdelta
	}
	var map = $('#world-map').vectorMap('get', 'mapObject')
	map.removeAllMarkers()
	map.addMarkers(markers)
	map.setFocus({scale:20000/cf.Distance, lat:from[0]+(to[0]-from[0])/2, lng:from[1]+(to[1]-from[1])/2, animate:true})
}
