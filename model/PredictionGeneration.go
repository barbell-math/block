package model;

import (
    "github.com/barbell-math/block/db"
    potSurf "github.com/barbell-math/block/model/potentialSurface"
    stateGen "github.com/barbell-math/block/model/stateGenerator"
)

//Given a set of values to use when making the prediction, the closest model
//state (in time) that is less than the current time and has the appropriate
//state generator and surface will be used to generate a prediction for intensity.
//All values besides Intensity need to be accurate in the training log argument.
//'current time' is defined by the 'DatePerformed' field of the training log arg.
func GeneratePrediction(
        c *db.DB,
        tl *db.TrainingLog,
        sg stateGen.StateGeneratorId,
        surf potSurf.PotentialSurfaceId) (db.Prediction,error) {
    rv:=db.Prediction{ TrainingLogID: tl.Id };
    if ms,err,found:=db.CustomReadQuery[db.ModelState](c,
        nearestModelStateToExerciseQuery(tl),[]any{
            tl.ExerciseID,
            tl.DatePerformed,
            sg,
            surf,
            tl.ClientID,
    }).Nth(0); err==nil && found {
        pred:=potSurf.CalculationsFromSurfaceId(
            potSurf.PotentialSurfaceId(ms.PotentialSurfaceID),
        );
        rv.TrainingLogID=tl.Id;
        rv.IntensityPred=pred.Intensity(ms,tl);
        rv.StateGeneratorID=ms.StateGeneratorID;
        rv.PotentialSurfaceID=ms.PotentialSurfaceID;
        return rv,nil;
    } else {
        return rv,err;
    }
}
