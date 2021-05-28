  var GoogleAuth;
  var SCOPE = 'openid';
  var gThemes={}
  function handleClientLoad() {
     var style = getComputedStyle(document.body);
     gThemes.primary = style.getPropertyValue('--primary');
     gThemes.secondary = style.getPropertyValue('--secondary');
     gThemes.success = style.getPropertyValue('--success');
     gThemes.info = style.getPropertyValue('--info');
     gThemes.warning = style.getPropertyValue('--warning');
     gThemes.danger = style.getPropertyValue('--danger');
     gThemes.light = style.getPropertyValue('--light');
     gThemes.dark = style.getPropertyValue('--dark');
     gapi.load('client:auth2', initClient);
  }

  function initClient() {
     gapi.client.init({
        'clientId': '502307674846-9hk8u2iaggriv00op22gk2iqoffmje9d.apps.googleusercontent.com',
	'scope': SCOPE,
	'ux_mode': 'redirect',
	'redirect_uri': 'https://flapyourarms.org/app/index.html' 
     }).then(function () {

      GoogleAuth = gapi.auth2.getAuthInstance();

      GoogleAuth.isSignedIn.listen(updateSigninStatus);

      var user = GoogleAuth.currentUser.get();
      setSigninStatus();

      $('#nav-signin').click(function() {
        handleAuthClick();
      });
      $('#nav-signout').click(function() {
        revokeAccess();
      });
      $('#nav-statistics').click(function() {
        waitCursorOn();
        showCharts();
      });  
      $('#nav-planning').click(function() {
	waitCursorOn();
	showPlanning();
      });
      $('#nav-history').click(function() {
       waitCursorOn();
       showHistory();
      });
      $('#nav-account').click(function() {
       waitCursorOn();
       showAccount();
      });

    });
  }

  function handleAuthClick() {
    if (GoogleAuth.isSignedIn.get()) {
      GoogleAuth.signOut();
    } else {
      GoogleAuth.signIn();
    }
  }

  function revokeAccess() {
    GoogleAuth.disconnect();
  }

var gPages=['signin','signout','statistics','planning','history','account']
var gPageTitles=['Welcome','Welcome','Statistics','Trip Planning','Flight History','Account Summary']

  function setSigninStatus(isSignedIn) {
    var user = GoogleAuth.currentUser.get();
    var isAuthorized = user.hasGrantedScopes(SCOPE);
    for (i in gPages) {
	if ((isAuthorized && (gPages[i]!='signin')) || (!isAuthorized && gPages[i]=='signin'))
	{
      		$('#nav-'+gPages[i]).removeClass("d-none")
	} else {
      		$('#nav-'+gPages[i]).addClass("d-none")
    	}
    }
    if (isAuthorized) {
	waitCursorOn()
	showAccount()
	updateEmail()
    } else {
	$("#useremail").text("")    
	navbarActive("signout")
    }
  }

  function updateSigninStatus(isSignedIn) {
    setSigninStatus();
  }

  var gPage
  function navbarActive(opt) {
	for (i in gPages) {
		if (opt == gPages[i]) {
			$('#nav-'+gPages[i]).css('fill','')
			$('#pg_'+gPages[i]).addClass("flapfadinelement")
			$('#pg_'+gPages[i]).removeClass("d-none")
			$("#pagetitle").text(gPageTitles[i])    
		} else {
			$('#nav-'+gPages[i]).css('fill','white')
			$('#pg_'+gPages[i]).addClass("d-none")
		}
	}
	gPage=opt
	waitCursorOff()
  }

  function waitCursorOn() {
	document.body.classList.add('inheritCursors');
	document.body.style.cursor = 'progress';
  }

  function waitCursorOff() {
	document.body.classList.remove('inheritCursors');
        document.body.style.cursor = 'unset';
  }


  var gBotBand=5
  var gBotNumber=0
  $('#changeModal').on('hidden.bs.modal', function (e) {

	  gBotNumber = $('#tNum').val()
	  gBotBand = $("#tBand").prop('selectedIndex')
	  historyInit=false
	  planningInit=false
	  accountInit=false
	  updateEmail()
	  $('#nav-account').trigger("click");
  })

  function updateSlider(val) {
     document.getElementById("botslidervalue").innerHTML = val;
  }
   
  function updateEmail() {
	//	$("#useremail").text(user.getBasicProfile().getEmail())
	$("#useremail").text($( "#tBand option:selected" ).text() + " Bot " +  gBotNumber.toString() )
  }
