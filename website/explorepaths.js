Cesium.Ion.defaultAccessToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJmNThhZDY0MC02OTRiLTRmNTgtOWM1Zi00OTU4MDc4MWQzYWQiLCJpZCI6MTgwOTIsInNjb3BlcyI6WyJhc3IiLCJnYyJdLCJpYXQiOjE1NzMyNDU0Mjd9.26QdibL7r1cKsLG3tmT2qH8WoBWXPvcsmrACrAyxW_c';
var viewer = new Cesium.Viewer('cesiumContainer',{animation:false,shouldAnimate:true,sceneMode:Cesium.SceneMode.SCENE2D});
showPaths(53812);

function showPaths(id)
{
    var promise = Cesium.IonResource.fromAssetId(id)
    .then(function (resource) {
	  return Cesium.KmlDataSource.load(resource, {
		camera: viewer.scene.camera,
		canvas: viewer.scene.canvas
	});
     })
     .then(function (dataSource) {
	dataSource.clock.clockRange=Cesium.ClockRange.CLAMPED;
	dataSource.clock.multiplier=86400*100
	return dataSource
     })
     .then(function (dataSource) {
	viewer.dataSources.removeAll();
	return dataSource
     })
     .then(function (dataSource) {
	return viewer.dataSources.add(dataSource);
     })
    .otherwise(function (error) {
	console.log(error);
     });
}

function selectDayRange(event) {
	var selectElement = event.target;
	var value = selectElement.value;
	showPaths(53811+parseInt(value,10))
}
