var historyInit=false
function showHistory() {
	if (!historyInit)
	{
		renderHistory()
       } else {
       		navbarActive('history');
       }
}

var gmapcreated=false
function renderHistory() {
	populateBoard()
	navbarActive('history');
	if (!gmapcreated) {
		$('#world-map').vectorMap({map: 'world_mill',zoomButtons : false, backgroundColor:'white',regionStyle: { initial: { fill: '#dc3545' }, hover: { fill: 'black' }} , 
					markerStyle:{initial: {fill: 'black',stroke: 'white',"fill-opacity": 1,"stroke-width": 1,"stroke-opacity": 1,r: 5}}});
		$('#world-map').vectorMap('get', 'mapObject').updateSize();
         	gmapcreated=true
	} else {
		var map = $('#world-map').vectorMap('get', 'mapObject')
		map.removeAllMarkers()
	}
        $('#flightList button').on('click', function (e) {
	   	e.preventDefault()
           	showFlightPath(parseInt($(this).attr("href"),10))
	})
        showFlightPath(0)
	historyInit=true
}

var flights=[]
function populateBoard() {
	var rpt=6
	flights = []
	$('#flightList').remove();
	var active=" active "
	var tf=0
	var rows="<div id='flightList' class='mx-auto'>"
	for (i in gTripHistory)  {
		for (j in gTripHistory[i].Journeys) {
			for (f in gTripHistory[i].Journeys[j].Flights) {
				if (tf%rpt==0) {
					rows+="<div class='list-group carousel-item justify-content-center" + active + "h6'>"
				}
				cf = gTripHistory[i].Journeys[j].Flights[f]
				flights.push(cf)
				start = moment.utc(cf.Start)
				end  = moment.utc(cf.End)
				rows += "<button class='list-group-item list-group-item-action border-white text-white bg-dark" + active + "' href='" + (flights.length-1).toString() + "'>" + start.format('YYYY-MM-DD') +  "&nbsp;&nbsp;" + start.format("hh:mm") + " " + gAirports[cf.From].Iata + "&nbsp;&nbsp;" + end.format("hh:mm") + " " + gAirports[cf.To].Iata + "</button>"
				tf++
				active=" "
				if (tf==rpt) {
					rows+="</div>"
					tf=0
				}
			}
		}
	}

	if (tf >0 && tf < rpt) {
		for (;tf < rpt; tf ++) {
			rows += "<a class='list-group-item border-white text-dark bg-dark' href='#'>&nbsp;</a>"
		}
		rows+="</div>"
	}
	rows += "</div>"
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

	$("#cityfrom").text(gAirports[cf.From].Cy);
	$("#cityto").text(gAirports[cf.To].Cy);
	$("#flightdist").text(Math.round(cf.Distance).toString()+" ");

}
