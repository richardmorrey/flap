var  accountInit=false
function showAccount() {
	if (!accountInit)
	{
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/account/id/'+id_token+"/b/"+ gBotBand + "/n/" + gBotNumber);
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {renderAccount((xhr.status == 200) ? xhr.responseText : "[]");};
	  xhr.send();

	  var xhr2 = new XMLHttpRequest();
	  xhr2.open('GET', '/user/v1/flighthistory/id/'+id_token+"/b/"+gBotBand + "/n/" + gBotNumber);
	  xhr2.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr2.onload = function() {renderAccountCharts((xhr2.status == 200) ? xhr2.responseText : "[]");};
	  xhr2.send();

       } else {
       		navbarActive('account')
       }
  }


var gTripHistory
var gFlights=[]
function renderAccountCharts(text) {
	gTripHistory=JSON.parse(text)
	for (i in gTripHistory)  {
		for (j in gTripHistory[i].Journeys) {
			for (f in gTripHistory[i].Journeys[j].Flights) {
				gFlights.push(gTripHistory[i].Journeys[j].Flights[f])
			}
		}
	}

	var dateLabels=[]
	var travelled=[]
	var footprint=[]
	var globalaverage=[]
	var ukaverage=[]
	var i=gFlights.length-1
	var currentDay = moment.utc(gFlights[i].Start)
	var now = moment()
	while (currentDay.isBefore(now)) {
		
		var dt = 0
		var fp = 0
		while (i >=0 && currentDay.isSame(moment.utc(gFlights[i].Start),'month')) {
			dt += gFlights[i].Distance
			fp += (moment.utc(gFlights[i].End).diff(moment.utc(gFlights[i].Start),"seconds")/3600)*.25
			i--
		}
		dateLabels.push(currentDay.toDate())
		travelled.push(dt)
		footprint.push(fp)
		ukaverage.push(13.4/12)
		globalaverage.push(5/12)
		currentDay = currentDay.add(1,"month")
	}

	var ctx = document.getElementById('mydistancechart').getContext('2d');
	var tdchart = new Chart(ctx, {
	  type: 'line',
	  data: {
		labels: dateLabels,		
		datasets: [
			        {data: movingAvg(travelled,3),borderColor: "#17a2b8",backgroundColor: "#17a2b8",fill:true,pointRadius:0}
		 	],
		},
	  options: {
		legend: { display:false},
		scales: {
			xAxes: [{type: "time", time: {unit: 'day', unitStepSize: 100,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "Distance Per Month (km)",fontColor: "black"}}]
			}
	           }
          });

	var ctx = document.getElementById('myfootprintchart').getContext('2d');
	var tdchart2 = new Chart(ctx, {
	  type: 'line',
	  data: {
		labels: dateLabels,		
		datasets: [
				{pointStyle:"line", label:"UK Avg (total)", data: ukaverage,borderColor: "black",fill:false,pointRadius:0,borderDash:[5,5]},
				{pointStyle:"line",label:"Global Avg (total)",data: globalaverage,borderColor:"black",fill:false,pointRadius:0,borderDash:[2,2]},
			        {pointStyle:"line",label:"You (flights only)", data: movingAvg(footprint,3),borderColor:"#17a2b8",backgroundColor:"#17a2b8",fill:true,pointRadius:0}
		],
		},
	  options: {
		legend: { labels: {usePointStyle:true}},
		scales: {
			xAxes: [{type: "time", time: {unit: 'day', unitStepSize: 100,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "CO2 per month (Tonnes)",fontColor: "black"}}]
			}
	           }
          });

}

var gAccount
function renderAccount(text) {
	gAccount = JSON.parse(text)
	var d=false,su=false,e=false

	switch (gAccount['Cleared'])
	{
		case 0:
			$('#grounded').addClass("badge-danger")
			$('#grounded').removeClass("badge-success")
			$('#grounded').removeClass("badge-secondary")
			$("#grounded").text("GROUNDED")
		break

		case 1:
			$('#grounded').addClass("badge-secondary")
			$('#grounded').removeClass("badge-success")
			$('#grounded').removeClass("badge-danger")
			$("#grounded").text("MID-TRIP")
		break

		default:
			$('#grounded').addClass("badge-success")
			$('#grounded').removeClass("badge-danger")
			$('#grounded').removeClass("badge-secondary")
			$("#grounded").text("CLEARED")
		break
	}

	$("#balance").text(Math.round(gAccount['Balance']).toString() + " km");
	if (gAccount['Balance'] > 0)
	{	
		$('#balance').addClass("badge-success")
		$('#balance').removeClass("badge-danger")
	}
	else
	{
		$('#balance').addClass("badge-danger")
		$('#balance').removeClass("badge-succes")
	}

	cd =  moment.utc(gAccount['ClearanceDate'])
	$("#clearancedate").text(cd.format("YYYY-MM-DD"))
	  
	
	var user = GoogleAuth.currentUser.get();
	var id_token = user.getAuthResponse().id_token;
        var xhr = new XMLHttpRequest();
	xhr.open('GET', '/user/v1/transactions/id/'+id_token+"/b/"+gBotBand + "/n/" + gBotNumber);
	xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhr.onload = function() {renderAccountTransactions((xhr.status == 200) ? xhr.responseText : "[]");};
	xhr.send();
}

function renderAccountTransactions(text)
{
	var balance= gAccount["Balance"]
	var history = JSON.parse(text)
	var ts=[]
	var bgs=[]
	var dateLabels=[]
	if (history.length > 0) {
		i = history.length-1
		var currentDay = moment.utc(history[i].Date)
		while (i >=0) {
			if (moment.utc(history[i].Date).isSame(currentDay,"day")) {
				ts.unshift([balance-history[i].Distance,balance])
				balance -= history[i].Distance
				bgs.unshift(history[i].Distance > 0 ? "black":"#dc3545")
				i--
			}
			else
			{
				ts.unshift([balance,balance])
				bgs.unshift("black")
			}
			dateLabels.unshift(currentDay.toDate())
			currentDay = currentDay.subtract(1,"day")
		}
	}

	var ctx = document.getElementById('mytransactionschart').getContext('2d');
	var tdchart = new Chart(ctx, {
	  type: 'bar',
	  data: {
		  labels: dateLabels,
		  datasets: [
    				{
      				data: ts,
      				backgroundColor: bgs,
    				}
  			]
		 },
	  options: {
		  legend: { display:false},
		  scales: {
			xAxes: [{offset:true,type: "time", time: {unit: 'day', unitStepSize: 5,round: 'day',displayFormats: {day: 'YYYY-MM-DD'}}}],
			yAxes: [{scaleLabel: {display: true,labelString: "Balance (km)",fontColor: "black"}}]
			}
	           }
          });
	navbarActive('account');
	accountInit=true

}
    /**
    * returns an array with moving average of the input array
    * @param array - the input array
    * @param count - the number of elements to include in the moving average calculation
    * @param qualifier - an optional function that will be called on each 
    *  value to determine whether it should be used
    */
    function movingAvg(array, count, qualifier){

        // calculate average for subarray
        var avg = function(array, qualifier){

            var sum = 0, count = 0, val;
            for (var i in array){
                val = array[i];
                if (!qualifier || qualifier(val)){
                    sum += val;
                    count++;
                }
            }

            return sum / count;
        };

        var result = [], val;

        // pad beginning of result with null values
        for (var i=0; i < count-1; i++)
            result.push(null);

        // calculate average for each subarray and add to result
        for (var i=0, len=array.length - count; i <= len; i++){

            val = avg(array.slice(i, i + count), qualifier);
            if (isNaN(val))
                result.push(null);
            else
                result.push(val);
        }

        return result;
    }
