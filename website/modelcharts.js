function day(d) {
	  return moment('2021-01-01').add(d, 'day').toDate();
};

var ctx = document.getElementById('goal1chart').getContext('2d');
var goal1 = new Chart(ctx, {
	  type: 'line',
	  data: {
	  	labels: [day(100),day(200),day(300),day(400),day(500),day(600),day(700),day(800),day(900),day(1000),day(1100),day(1200),day(1300),day(1400),day(1500),day(1600),day(1700),day(1800),day(1900),day(2000)],
		
		 datasets: [{
				label: 'Daily Total',
				data: [
					null,
					12030027,
					11443125,
					10884854,
					10353814,
					9848676,
					9368182,
					8911129,
					8476373,
					8062828,
					7669459,
					7295275,
					6939344,
					6600775,
					6278724,
					5972382,
					5680985,
					5403802,
					5140140,
					4889343
					],
				borderColor: "rgba(200,0,0,1)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false,
			 	pointRadius:0
			   },
			   {
				label: 'Distance Travelled',
				data: [
					12330347,
					12548112,
					12184289,
					11717813,
					11310505,
					10929673,
					10422789,
					9955911,
					9533275,
					9108709,
					8685640,
					8303961,
					7951723,
					7530432,
					7177688,
					6821608,
					6496365,
					6194479,
					5920555,
					5674659
					],
				  borderColor: "rgba(0,0,255,1)",
				  backgroundColor: "rgba(0,0,255,0.6)",
				  fill:false
				}]
			  },
	options: {
		legend: {
			labels: {usePointStyle:true}
		},
		scales: {
			xAxes: [{
				      type: "time",
				      time: {
					              unit: 'day',
					              unitStepSize: 100,
					              round: 'day',
					              displayFormats: {
							                day: 'YYYY-MM-DD'
							              }
					            }
				    }],
			    yAxes: [{
				          scaleLabel: {
						          display: true,
						          labelString: "Distance Per Day (km)",
						          fontColor: "black"
						        }
				        }]
			  }
	}
}
);
var ctx = document.getElementById('goal2chart').getContext('2d');
var goal2 = new Chart(ctx, {
	  type: 'bar',
	  data: {
	  	labels: [day(100),day(200),day(300),day(400),day(500),day(600),day(700),day(800),day(900),day(1000),day(1100),day(1200),day(1300),day(1400),day(1500),day(1600),day(1700),day(1800),day(1900),day(2000)],
		
		 datasets: [{
				label: 'Frequent Flyers',
				data: [
					3715.273107,
					3718.2683,
					3515.983747,
					3309.589727,
					3105.16958,
					2885.769073,
					2681.97386,
					2459.641587,
					2268.18758,
					2092.76596,
					1911.060687,
					1756.62384,
					1628.28076,
					1469.81358,
					1371.75436,
					1244.238953,
					1154.388193,
					1047.518327,
					969.559127,
					920.631013
					],
				backgroundColor: "rgba(255,165,0,0.5)",
			 	borderColor: "rgba(255,165,0,1)",
			 	borderWidth:1
			   },
			   {
				label: 'Infrequent Flyers',
				data: [
					794.992725,
					820.08354,
					812.978161,
					794.521087,
					782.676612,
					776.59056,
					752.92104,
					737.229284,
					721.293381,
					702.301192,
					684.594005,
					666.944184,
					648.153184,
					626.554387,
					602.359686,
					582.970641,
					560.562774,
					543.906122,
					525.437308,
					505.142756
					],
				  backgroundColor: "rgba(255,255,0,0.5)",
			          borderColor: "rgba(255,255,0,1)",
				  borderWidth: 1
	}]
			  },
	options: {
		scales: {
			xAxes: [{
				      offset:true,
				      type: "time",
				      time: {
					              unit: 'day',
					              unitStepSize: 100,
					              round: 'day',
					              displayFormats: {
							                day: 'YYYY-MM-DD'
							              }
					            },
				      gridLines: {
					      display:false
				      }

				    }],
			    yAxes: [{
				          scaleLabel: {
						          display: true,
						          labelString: "Distance Per 100 Days (km)",
						          fontColor: "black"
						        },
				          ticks: {
						  beginAtZero:true
					  	}
				        }]
			  }
	}
}
);
var ctx = document.getElementById('goal3chart').getContext('2d');
var goal3 = new Chart(ctx, {
	  type: 'line',
	  data: {
	  	labels: [day(100),day(200),day(300),day(400),day(500),day(600),day(700),day(800),day(900),day(1000),day(1100),day(1200),day(1300),day(1400),day(1500),day(1600),day(1700),day(1800),day(1900),day(2000)],
		
		 datasets: [{
				label: 'Frequent Flyers',
				data: [
					0,
					1.958017,
					4.849405,
					7.684052,
					10.812927,
					14.365406,
					18.39783,
					22.213644,
					26.256476,
					30.304115,
					34.08941,
					37.930871,
					41.439207,
					45.306004,
					48.218464,
					51.615124,
					54.679067,
					57.664198,
					60.478904,
					62.371811
					],
				backgroundColor: "rgba(255,165,0,0.5)",
			 	borderColor: "rgba(255,165,0,1)",
				fill:false,
			 	pointRadius:0
			   },
			   {
				label: 'Infrequent Flyers',
				data: [
					0,
					0.42893,
					1.0716,
					1.671664,
					2.340001,
					3.248634,
					4.231149,
					5.336774,
					6.536206,
					7.741418,
					9.017696,
					10.346857,
					11.888425,
					13.380424,
					14.802662,
					16.636672,
					18.362407,
					19.937598,
					21.58502,
					23.58651
					],
				  backgroundColor: "rgba(255,255,0,0.5)",
			          borderColor: "rgba(255,255,0,1)",
				  fill:false,
				  pointRadius:0
				}]
			  },
	options: {
		legend: {
			labels: {usePointStyle:true}
		},
		scales: {
			xAxes: [{
				      type: "time",
				      time: {
					              unit: 'day',
					              unitStepSize: 100,
					              round: 'day',
					              displayFormats: {
							                day: 'YYYY-MM-DD'
							              }
					            },
				}],
			    yAxes: [{
				          scaleLabel: {
						          display: true,
						          labelString: "Trips Rejected (%)",
						          fontColor: "black"
						        }
				        }]
			  }
	}
}
);

