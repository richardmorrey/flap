function day(d) {
	  return moment('2021-01-01').add(d, 'day').toDate();
};

function buildDates(from,to) {
	var dates= []
	while (from <= to) {
	   dates.push(day(from))
	   from +=10
	}
	return dates
}

function showCharts() {
	  navbarActive('statistics')
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/dailystats/id/'+id_token);
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   renderCharts(xhr.responseText);
			 };
	  xhr.send();
  }

function renderCharts(text) {
	raw = JSON.parse(text)
	var dateLabels=[]
	var travelled=[]
	var usertravelled=[]
	var flights=[]
	var userflights=[]
	var da=[]
	var share=[]
	for (i in raw) {
		dateLabels.push(moment(1000*raw[i].Date).toDate())
		travelled.push(raw[i].Travelled)
		flights.push(raw[i].Flights)
		da.push(raw[i].DailyTotal)
		usertravelled.push(raw[i].Travelled/10000)
		userflights.push(raw[i].Flights/10000)
		share.push(raw[i].share)
	}

	var ctx = document.getElementById('distancechart').getContext('2d');
	var tdchart = new Chart(ctx, {
	  type: 'line',
	  data: {
		labels: dateLabels,		
		datasets: [
				{pointStyle:"line",label: 'Daily Allowance',data: da,borderColor: "rgba(1,1,1)",backgroundColor: "rgba(1,1,1)",borderDash: [10,10],fill:false,pointRadius:0},
			        {pointStyle:"line",label: 'Distance Travelled',data: movingAvg(travelled,30),borderColor: "#dc3545",fill:false,pointRadius:0}
		 	]
		},
	  options: {
		legend: { labels: {usePointStyle:true}},
		scales: {
			xAxes: [{type: "time", time: {unit: 'day', unitStepSize: 100,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "Distance Per Day (km)",fontColor: "black"}}]
			}
	           }
          });
	var ctx2 = document.getElementById('userdistancechart').getContext('2d');
	var dchart = new Chart(ctx2, {
	  type: 'line',
	  data: {
		labels: dateLabels,		
		datasets: [
				{pointStyle:"line",label: 'Daily Share',data: share,borderColor: "rgba(1,1,1)",backgroundColor: "rgba(1,1,1)",borderDash: [10,10],fill:false,pointRadius:0},
			        {pointStyle:"line",label: 'Distance Travelled',data: movingAvg(usertravelled,30),borderColor: "#dc3545",fill:false,pointRadius:0}
		 	]
		},
	  options: {
		legend: { labels: {usePointStyle:true}},
		scales: {
			xAxes: [{type: "time", time: {unit: 'day', unitStepSize: 100,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "Distance Per Day (km)",fontColor: "black"}}]
			}
	           }
          });	
	var ctx3 = document.getElementById('flightschart').getContext('2d');
	var dchart = new Chart(ctx3, {
	  type: 'line',
	  data: {
		labels: dateLabels,		
		datasets: [
			        {pointStyle:"line",label: 'Flights',data: movingAvg(flights,30),borderColor: "#dc3545",fill:false,pointRadius:0}
		 	]
		},
	  options: {
		legend: { labels: {usePointStyle:true}},
		scales: {
			xAxes: [{type: "time", time: {unit: 'day', unitStepSize: 100,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "Flights",fontColor: "black"}}]
			}
	           }
          });

}

