# Tools for using Kata Containers with VFIO devices (esp. SR-IOV)

## Content

/containers directory has some samble containers with basic testcases
for VFIO devices.

/deployments directory has some useful deployment yaml files

/scripts has some useful scripts

## Podman

[Instructions for using VFIO in Kata containers via Podman](podman.md)

## Debugging

To debug the Kata andbox VM, you can add a console by adding
`agent.debug_console` to the `kernel_params` variable in
`configuration.toml`.

You can then connect to that debug console of a Kata container by using:
```
./scripts/kata-console [<container's UUID>]
```

If the UUID is omitted, it will connect to the last Kata container
started (very useful when you're only running one container at a time.

## Contact

David Gibson <david@gibson.dropbear.id.au>
Pradipta Kr. Banerjee <pradipta.banerjee@gmail.com>
