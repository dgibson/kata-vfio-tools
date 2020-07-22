# FIXME - needs updating

List your PCI devices' vendor/device ids:

    lspci -nn | grep Ethernet

Tweak the configMap to match your devices accordingly

Make sure your devices are bound to `vfio` driver.

Deploy the configMap

    kubectl apply -f deployments/configMap.yaml

Deploy the device plugin

    kubectl apply -f deployments/sriov-dp.yaml

Make sure the DP has detected your devices and created the pools

    kubectl get node $HOSTNAME -o json | jq '.status.allocatable'
    {
      "cpu": "28",
      "ephemeral-storage": "66100978374",
      "hugepages-1Gi": "0",
      "hugepages-2Mi": "8Gi",
      "intel.com/700PF": "1",
      "intel.com/700VF": "4",
      "memory": "39369372Ki",
     "pods": "110"
   }

Deploy the testing pod

    kubectl apply -f deployments/fedora-sriov.yaml


