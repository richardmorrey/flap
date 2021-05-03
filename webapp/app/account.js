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

	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr2 = new XMLHttpRequest();
	  xhr2.open('GET', '/user/v1/flighthistory/id/'+id_token+"/b/"+gBotBand + "/n/" + gBotNumber);
	  xhr2.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr2.onload = function() {renderAccountCharts((xhr2.status == 200) ? xhr2.responseText : "[]");};
	  xhr2.send();

	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr3 = new XMLHttpRequest();
	  xhr3.open('GET', '/user/v1/transactions/id/'+id_token+"/b/"+gBotBand + "/n/" + gBotNumber);
	  xhr3.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr3.onload = function() {renderAccountTransactions((xhr2.status == 200) ? xhr3.responseText : "[]");};
	  xhr3.send();

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
			        {data: movingAvg(travelled,3),borderColor: "black",backgroundColor: "#dc3545",fill:true,pointRadius:0}
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
			        {pointStyle:"line",label:"You (flights only)", data: movingAvg(footprint,3),borderColor:"black",backgroundColor:"#dc3545",fill:true,pointRadius:0},
				{pointStyle:"line", label:"UK Avg (total)", data: ukaverage,borderColor: "black",fill:false,pointRadius:0,borderDash:[5,5]},
				{pointStyle:"line",label:"Global Avg (total)",data: globalaverage,borderColor:"black",fill:false,pointRadius:0,borderDash:[2,2]}
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

function renderAccount(text) {
	var account = JSON.parse(text)
	var d=false,su=false,e=false

	switch (account['Cleared'])
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

	$("#balance").text(Math.round(account['Balance']).toString() + " km");
	if (account['Balance'] > 0)
	{	
		$('#balance').addClass("badge-success")
		$('#balance').removeClass("badge-danger")
	}
	else
	{
		$('#balance').addClass("badge-danger")
		$('#balance').removeClass("badge-succes")
	}

	cd =  moment.utc(account['ClearanceDate'])
	$("#clearancedate").text(cd.format("YYYY-MM-DD"))

	navbarActive('account');
	accountInit=true
}

function renderAccountTransactions(text)
{
	//alert(text)
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
