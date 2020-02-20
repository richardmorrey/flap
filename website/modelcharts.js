function day(d) {
	  return moment('2021-01-01').add(d, 'day').toDate();
};

var ctx = document.getElementById('goal1chart').getContext('2d');
var goal1 = new Chart(ctx, {
	  type: 'line',
	  data: {
	  	labels: [day(100),day(200),day(300),day(400),day(500),day(600),day(700),day(800),day(900),day(1000),day(1100),day(1200),day(1300),day(1400),day(1500),day(1600),day(1700),day(1800),day(1900),day(2000)],
		
		 datasets: [{
				label: 'Daily Allowance',
				data: [
					null,
					13095997,
					12457093,
					11849349,
					11271258,
					10721370,
					10198306,
					9700761,
					9227485,
					8777299,
					8349071,
					7941734,
					7554268,
					7185702,
					6835117,
					6501636,
					6184421,
					5882682,
					5595661,
					5322640
					],
				borderColor: "rgba(200,0,0,1)",
				backgroundColor: "rgba(200,0,0,0.6)",
				fill:false,
			 	pointRadius:0
			   },
			   {
				label: 'Distance Travelled',
				data: [
					13429637,
					13413465,
					13433483,
					13498419,
					13353718,
					13247649,
					12985126,
					12620920,
					12302526,
					11880445,
					11433243,
					11053795,
					10652597,
					10327351,
					10002187,
					9670491,
					9413617,
					9125704,
					8914816,
					8675271
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
					4169.240087,
					4187.631807,
					4173.579293,
					4199.555687,
					4147.476573,
					4060.089053,
					3913.362447,
					3656.81048,
					3440.489193,
					3193.317967,
					2900.2705,
					2635.032133,
					2365.031633,
					2157.904153,
					1983.9955,
					1816.929827,
					1670.495467,
					1562.266587,
					1452.1618,
					1357.768847,
					],
				backgroundColor: "rgba(255,165,0,0.5)",
			 	borderColor: "rgba(255,165,0,1)",
			 	borderWidth:1
			   },
			   {
				label: 'Infrequent Flyers',
				data: [
					844.209142,
					839.06086,
					843.895813,
					846.951322,
					839.118128,
					842.060753,
					837.068586,
					839.494702,
					840.210919,
					834.172828,
					833.274988,
					835.440907,
					835.888256,
					834.175911,
					826.611068,
					817.070168,
					812.691082,
					797.91814,
					792.538078,
					781.013931
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
					0.01133,
					0.22502,
					0.932938,
					2.55058,
					5.517655,
					10.212946,
					15.850274,
					21.871959,
					28.646494,
					34.937439,
					41.187517,
					46.440642,
					50.935349,
					54.697761,
					58.545086,
					61.196442,
					63.844282,
					66.315698,
					68.393993
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
					0.000586,
					0.000587,
					0.002936,
					0.010026,
					0.024175,
					0.054287,
					0.115999,
					0.21057,
					0.383726,
					0.592183,
					0.86264,
					1.2497,
					1.753291,
					2.331946,
					3.102432,
					3.935093,
					4.720542,
					5.917639,
					7.189397
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
						          labelString: "Trips Cancelled (%)",
						          fontColor: "black"
						        }
				        }]
			  }
	}
}
);

