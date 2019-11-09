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
				borderColor: "rgba(200,0,0,1)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false,
			 	pointRadius:0
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
					6676.16,
					6051.84,
					5864.89,
					4312.1,
					3890.79,
					3957.75,
					3226.31,
					2986.61,
					2378.99,
					1678.96,
					2661.65,
					1940.08,
					1642.78,
					951.92,
					828.33,
					1222.98,
					1622.33,
					1029.04,
					822.59,
					922.33
					],
				backgroundColor: "rgba(255,165,0,0.5)",
			 	borderColor: "rgba(255,165,0,1)",
			 	borderWidth:1
			   },
			   {
				label: 'Infrequent Flyers',
				data: [
					594.937778,
					851.996667,
					817.966667,
					754.246667,
					847.561111,
					727.304444,
					668.293333,
					815.804444,
					775.311111,
					678.823333,
					577.228889,
					597.56,
					615.186667,
					607.738889,
					478.133333,
					634.127778,
					523.807778,
					469.633333,
					632.584444,
					392.498889
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
					0.37594,
					0.83682,
					0.350877,
					0.921659,
					1.190476,
					7.30897,
					5.714286,
					15.384615,
					21.338912,
					23.004695,
					28.365385,
					33.179724,
					37.681159,
					42.79476,
					48.514851,
					45.5,
					54.929577,
					52.717391,
					66.875
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
					0.320513,
					0.314465,
					0.294985,
					0.289017,
					0.928793,
					0.900901,
					3.459119,
					2.605863,
					4.807692,
					5.18732,
					5.993691,
					10.469314,
					11.428571,
					12,
					13.05638,
					16.044776,
					17.518248,
					21.785714,
					20.472441
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

