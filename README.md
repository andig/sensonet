# sensonet modules as extension for evcc 

This is a fork of evcc.io to add an extension for Vaillant heat pumps with the sensoNET control module.

## Features

- Adds a charger module sensonet und a vehicle module sensonet_vehicle to allow PV optimised heating with a Vaillant heat pump
- The sensonet module connects to the myVaillant portal to get system information from the heat pump and to start quick modes.
- As the sensonet charger module supports two different charge modes ('Hotwater Boost' and 'Heating Quick Veto'), small additions were made to the
  module loadpoint.go and the vue file Vehicle.vue to display the current charge mode. The usage of these two changes is optional. 

## Getting Started

You'll find everything you need in our [documentation](https://docs.evcc.io/).
