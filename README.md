# sensonet modules as extension for evcc 

This is a fork of evcc-io/evcc to add an extension for Vaillant heat pumps with the sensoNET (VR921) control module.

This fork of evcc should only be used if you have a Vaillant heat pump and want to use evcc for PV optimised hotwater production or heating.
All other persons should visit github.com/evcc-io/evcc. There you can also find all general documentation about evcc.

## Features

- Adds a charger module sensonet und a vehicle module sensonet_vehicle to allow PV optimised heating with a Vaillant heat pump
- The sensonet module connects to the myVaillant portal to get system information from the heat pump and to start quick modes.
- As the sensonet charger module supports two different charge modes ('Hotwater Boost' and 'Heating Quick Veto'), small additions were made to the
  module loadpoint.go and the vue file Vehicle.vue to display the current charge mode. The usage of these two changes is optional. 

## How it works

To use this extension a charger section for 'sensonet' and a vehicle section for 'sensonet_vehicle' have to be added to evcc.yaml. If an energy meter
like Shelly 3EM is present for the Vaillant heatpump, then a meter section has to be added to evcc.yaml as well.
The sensonet charger, the sensonet_vehicle vehicle and the energy meter have to be combinated as a loadpoint in the yaml file.

The sensonet charge module initiates a http API interface to the myVaillant portal and keeps this connection active by regular refresh token calls.
A json system report is obtained about every two minutes using a http GET call. This report is analysed to extract information about heating zones, operating modes, current temperatures and temperature setpoints.
The sensonet_vehicle module is used to present some information like SoC (=current temperature) and TargetSoC (=temperature setpoint) to the loadpoint module of evcc.

If the evcc loadpoint module signals to the sensonet charger module that a charging session should be started, then the sensonet module does the following:
   (a) If ('pvusestrategy='hotwater') or (pvusestrategy='hotwater_than_heating' and current hotwater temperature is more than 5°C below the setpoint), then
         a hotwater boost is initiated via http POST request.
   (b) If ('pvusestrategy='heating') or (pvusestrategy='hotwater_than_heating' and current hotwater temperature is less than 5°C below the setpoint), then
         a heating zone quick veto is initiated via http POST request for the zone given by the heatingzone parameter in the evcc.yaml file (default=0). 
        The duration of the veto is 30 minutes and the veto_setpoint= normal_temperature_setpoint + heatingtemperatureoffset (default=2).

Via the json system reports, the sensonet module notices when the hotwater boost or the zone quick veto ends and changes the enabled() status to false to report
this to the loadpoint module of evcc.
If the loadpoint module tells the sensonet module to stop a charging session, then the sensonet module sends a "cancel hotwater boost" or "cancel zone quick veto" to the myVaillant portal via http DELETE request.

It happens, that it takes a few minutes before the json system report from the myVaillant portal reflects an initiated or canceled boost or quick veto. Don't worry!

## Warning

This extension is still unstable and in first tests.
Feedback of beta testers is welcome.
