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
       } else {
       		navbarActive('account')
       }
  }

function renderAccount(text) {
	alert(text)
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

