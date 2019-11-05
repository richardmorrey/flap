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
					7884,
					15616,
					14816,
					14027,
					13313,
					12613,
					11932,
					11325,
					10725,
					10129,
					9603,
					9103,
					8603,
					8107,
					7682,
					7281,
					6881,
					6482,
					6086,
					5761
					],
				borderColor: "rgba(200,0,0,0.6)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false
			   },
			   {
				label: 'Distance Travelled',
				data: [
					16063,
					14982,
					13989,
					14569,
					13575,
					14416,
					13315,
					12635,
					12986,
					12111,
					10356,
					9230,
					10129,
					9998,
					11459,
					6463,
					7523,
					6940,
					7348,
					7871
					],
				  borderColor: "rgba(255,153,0,0.6)",
				  backgroundColor: "rgba(255,153,0,0.6)",
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
				          gridLines: {
						          color: "black",
						          borderDash: [2, 5],
						        },
				          scaleLabel: {
						          display: true,
						          labelString: "Distance (Kilometres)",
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
					0.417778,
					0.348889,
					0.397778,
					0.38,
					0.367778,
					0.372222,
					0.373333,
					0.354444,
					0.397778,
					0.371111,
					0.392222,
					0.354444,
					0.32,
					0.368889,
					0.352222,
					0.305556,
					0.296667,
					0.318889,
					0.295556,
					0.297778
					],
				borderColor: "rgba(200,0,0,0.6)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false
			   },
			   {
				label: 'Infrequent Travellers',
				data: [
					],
				  borderColor: "rgba(255,153,0,0.6)",
				  backgroundColor: "rgba(255,153,0,0.6)",
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
				          gridLines: {
						          color: "black",
						          borderDash: [2, 5],
						        },
				          scaleLabel: {
						          display: true,
						          labelString: "Distance (Kilometres)",
						          fontColor: "black"
						        }
				        }]
			  }
	}
}
);
var ctx = document.getElementById('goal3chart').getContext('2d');
var goal3 = new Chart(ctx, {
	  type: 'bar',
	  data: {
	  	labels: [day(100),day(200),day(300),day(400),day(500),day(600),day(700),day(800),day(900),day(1000),day(1100),day(1200),day(1300),day(1400),day(1500),day(1600),day(1700),day(1800),day(1900),day(2000)],
		
		 datasets: [{
				label: 'Daily Total',
				data: [
					],
				borderColor: "rgba(200,0,0,0.6)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false
			   },
			   {
				label: 'Distance Travelled',
				data: [
					],
				  borderColor: "rgba(255,153,0,0.6)",
				  backgroundColor: "rgba(255,153,0,0.6)",
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
				          gridLines: {
						          color: "black",
						          borderDash: [2, 5],
						        },
				          scaleLabel: {
						          display: true,
						          labelString: "Distance (Kilometres)",
						          fontColor: "black"
						        }
				        }]
			  }
	}
}
);

