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
var gDistChart
var gFootprintChart
var gCountryChart
function renderAccountCharts(text) {

	if (gDistChart) {
		gDistChart.destroy()
		gFootprintChart.destroy()
		gCountryChart.destroy()
	}

	gTripHistory=JSON.parse(text)
	for (i in gTripHistory)  {
		for (j in gTripHistory[i].Journeys) {
			for (f in gTripHistory[i].Journeys[j].Flights) {
				gFlights.push(gTripHistory[i].Journeys[j].Flights[f])
			}
		}
	}

	var distLabels=[]
	var footLabels=[]
	var travelled=[]
	var footprint=[]
	var globalaverage=[]
	var ukaverage=[]
	var i=gFlights.length-1
	var currentDay = moment.utc(gFlights[i].Start)
	var now = moment()
	var byCountry={}
	while (currentDay.isBefore(now)) {
		
		var dt = 0
		var fp = 0
		while (i >=0 && currentDay.isSame(moment.utc(gFlights[i].Start),'month')) {
			dt += gFlights[i].Distance
			co = [gAirports[gFlights[i].To].Co]
			if (byCountry[co] == null) {
				byCountry[co] = gFlights[i].Distance 
			}
			else
			{
				byCountry[co] += gFlights[i].Distance
			}
			fp += (moment.utc(gFlights[i].End).diff(moment.utc(gFlights[i].Start),"seconds")/3600)*.25
			i--
		}
		distLabels.push(currentDay.toDate())
		footLabels.push(currentDay.toDate())
		travelled.push(dt)
		footprint.push(fp)
		ukaverage.push(13.4/12)
		globalaverage.push(5/12)
		currentDay = currentDay.add(1,"month")
	}
	byCountry = removeHighest(byCountry)
	
	var ctx = document.getElementById('mydistancechart').getContext('2d');
	gDistChart = new Chart(ctx, {
	  type: 'line',
	  data: {
		labels: distLabels,		
		datasets: [
			        {data: movingAvg(travelled,3),borderColor: "black",borderWidth:1,backgroundColor: "#17a2b8",fill:true,pointRadius:0}
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

	var ctx2 = document.getElementById('myfootprintchart').getContext('2d');
	gFootprintChart = new Chart(ctx2, {
	  type: 'line',
	  data: {
		labels: footLabels,		
		datasets: [
				{pointStyle:"line", label:"UK Avg (total)", data: ukaverage,borderColor: "black",fill:false,pointRadius:0,borderDash:[5,5]},
				{pointStyle:"line",label:"Global Avg (total)",data: globalaverage,borderColor:"black",fill:false,pointRadius:0,borderDash:[2,2]},
			        {pointStyle:"line",label:"You (flights only)",data: movingAvg(footprint,3),borderWidth:1,borderColor:"black",backgroundColor:"#17a2b8",fill:true,pointRadius:0}
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
	
	var ctx3 = document.getElementById('mycountrychart').getContext('2d');
	gCountryChart = new Chart(ctx3, {
	  type: 'doughnut',
	  data: {
		  labels: Object.keys(byCountry),
		  datasets: [
			{
				data:Object.values(byCountry),
				backgroundColor:Object.values(gThemes),
				borderColor:"black",
				borderWidth:"1"
			}
		],
		},
	  options: {
		legend: { labels: {usePointStyle:true}},
	           },
		 plugins: {
      			legend: {
        			position: 'bottom',
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
			$('#grounded').removeClass("badge-warning")
			$("#grounded").text("grounded")
		break

		case 1:
			$('#grounded').addClass("badge-warning")
			$('#grounded').removeClass("badge-success")
			$('#grounded').removeClass("badge-danger")
			$("#grounded").text("mid-trip")
		break

		default:
			$('#grounded').addClass("badge-success")
			$('#grounded').removeClass("badge-danger")
			$('#grounded').removeClass("badge-warning")
			$("#grounded").text("cleared")
		break
	}

	$("#balance").text(Math.round(gAccount['Balance']).toString());
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
	if (cd.isAfter(moment())) {
		$("#clearancedate").text(cd.format("YYYY-MM-DD"))
		$('#clearancedate').addClass("badge-danger")
		$('#clearancedate').removeClass("badge-success")
	} else {
		$("#clearancedate").text("now")
		$('#clearancedate').removeClass("badge-danger")
		$('#clearancedate').addClass("badge-success")
	}
	
	var user = GoogleAuth.currentUser.get();
	var id_token = user.getAuthResponse().id_token;
        var xhr = new XMLHttpRequest();
	xhr.open('GET', '/user/v1/transactions/id/'+id_token+"/b/"+gBotBand + "/n/" + gBotNumber);
	xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhr.onload = function() {renderAccountTransactions((xhr.status == 200) ? xhr.responseText : "[]");};
	xhr.send();
}

var gTransactionChart
function renderAccountTransactions(text)
{
	if (gTransactionChart) {gTransactionChart.destroy()}
	var balance= gAccount["Balance"]
	var history = JSON.parse(text)
	var ts=[]
	var bgs=[]
	var dateLabels=[]
	var td=0
	if (history.length > 0) {
		i = history.length-1
		var currentDay = moment.utc(history[i].Date)
		while (i >=0) {
			var t = 0
			while (i>=0 && moment.utc(history[i].Date).isSame(currentDay,"day")) {
				t += history[i].Distance
				i--
			}
			if (t != 0) {
				ts.unshift([balance-t,balance])
				balance -= t
				bgs.unshift(t> 0 ? "black":"#dc3545")
				t=0
				dateLabels.unshift(currentDay.toDate())
			}
			currentDay = currentDay.subtract(1,"day")
			td += 1
		}
	}
	var ctx = document.getElementById('mytransactionschart').getContext('2d');
	gTransactionChart = new Chart(ctx, {
	  type: 'bar',
	  data: {
		  labels: dateLabels,
		  datasets: [
    				{
      				data: ts,
      				backgroundColor: bgs,
				borderColor:"black",
				borderWidth: td > 30 ? 0 : 1
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
    function movingAvg(array,count, qualifier){

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

    function removeHighest(obj) {
	keysdesc = Object.keys(obj).sort(function(a,b){return obj[b]-obj[a]})
	//delete obj[keysdesc[0]]
	var vOthers=0
	for (i = 5 ; i < keysdesc.length; i++) {
		vOthers += obj[keysdesc[i]]
		delete obj[keysdesc[i]]
	}
	obj["Others"]=vOthers
	return obj
    }
