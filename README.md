# Latimer
Latimer is a tool for orchestrating installation of SW packages in kubernetes systems using HELM.

## Motivation
Helm by itself is great to install individuall sw packages, but in installing a large system composed by many Helm charts with installation dependencies across them, there is no tool out there which does this 100%.   Terraform has a nice HELM plugin which can be used to express dependencies, but it frequently fails in waiting for an installation to fully complete and pods to be in fully READY state.

Latimer intends to be about predictable/ordered installations, giving full traceability to assess when things go wrong and also provide clean and ordered uninstallation.
