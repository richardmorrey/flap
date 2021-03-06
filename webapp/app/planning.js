var  planningInit=false
function showPlanning() {
	if (!planningInit)
	{
	  var user = GoogleAuth.currentUser.get();
	  var id_token = user.getAuthResponse().id_token;
	  var xhr = new XMLHttpRequest();
	  xhr.open('GET', '/user/v1/promises/id/'+id_token+"/b/5/n/1");
	  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	  xhr.onload = function() {
				   renderPlanning(xhr.responseText);
			 };
	  xhr.send();
       } else {
       		navbarActive('planning')
       }
  }

var events=[]
function renderPlanning(text) {
	var promises = JSON.parse(text)
	var stacksize=0
	for (i in promises) {
		if (promises[i].Stacked == 0) {
			for (y=i; y >= Math.max(0, i-stacksize);y--) {
			    promises[y].level=stacksize
			}
			stacksize=0
		}
	}

	for (i in promises)  {
		events.push({level:promises[i].level,startDate: new Date(promises[i].TripStart.substr(0,10)), endDate: new Date(promises[i].TripEnd.substr(0,10))})
	}

	var cal = new Calendar('#calendar',{
		dataSource: events,
		customDataSourceRenderer: dsRender,
		style: "custom"
		})
	navbarActive('planning');
	planningInit=true
}

function dsRender(element, date, events) {
	var bgs=['grey','dimgrey','black']
	$(element).css('background-color', bgs[events[0].level]);
	$(element).css('color',"white");
        $(element).css('border-radius', '0px');
}

